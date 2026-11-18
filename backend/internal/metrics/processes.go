package metrics

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// HostProcess es un grupo de procesos del host (fuera de Docker) con el mismo nombre.
type HostProcess struct {
	Name     string `json:"name"`
	Count    int    `json:"count"`
	MemBytes uint64 `json:"mem_bytes"`
}

const maxHostProcesses = 8

// ReadHostProcesses responde "en que se va la RAM fuera de Docker": recorre
// procPath/[pid] (con pid:host el dashboard ve todos los procesos del host, solo
// lectura), salta los que viven en un cgroup de contenedor -- esos ya se cuentan por
// contenedor via docker stats -- y agrupa por nombre.
//
// Se usa RssAnon (memoria propia del proceso), no VmRSS: VmRSS incluye librerias y
// archivos mapeados compartidos que se contarian varias veces y que el kernel libera
// bajo presion, igual que el page cache que tampoco cuenta como "usado".
func ReadHostProcesses(procPath string) ([]HostProcess, error) {
	entries, err := os.ReadDir(procPath)
	if err != nil {
		return nil, err
	}

	byName := map[string]*HostProcess{}
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		dir := filepath.Join(procPath, e.Name())
		if inContainer(dir) {
			continue
		}
		name, anonKB, ok := readStatus(dir)
		if !ok || anonKB == 0 { // hilos del kernel no tienen memoria de usuario
			continue
		}
		p := byName[name]
		if p == nil {
			p = &HostProcess{Name: name}
			byName[name] = p
		}
		p.Count++
		p.MemBytes += anonKB * 1024
	}

	out := make([]HostProcess, 0, len(byName))
	for _, p := range byName {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MemBytes > out[j].MemBytes })

	if len(out) > maxHostProcesses {
		rest := HostProcess{Name: "otros"}
		for _, p := range out[maxHostProcesses:] {
			rest.Count += p.Count
			rest.MemBytes += p.MemBytes
		}
		out = append(out[:maxHostProcesses], rest)
	}
	return out, nil
}

// inContainer: con cgroup v2 + systemd, un proceso de contenedor vive en
// ".../docker-<id>.scope". dockerd/containerd y los shims NO (son del host).
func inContainer(dir string) bool {
	b, err := os.ReadFile(filepath.Join(dir, "cgroup"))
	if err != nil {
		return true // proceso que ya termino o ilegible: no adivinar, se omite
	}
	s := string(b)
	return strings.Contains(s, "/docker-") || strings.Contains(s, "/docker/")
}

func readStatus(dir string) (name string, anonKB uint64, ok bool) {
	f, err := os.Open(filepath.Join(dir, "status"))
	if err != nil {
		return "", 0, false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, val, found := strings.Cut(sc.Text(), ":")
		if !found {
			continue
		}
		val = strings.TrimSpace(val)
		switch key {
		case "Name":
			name = val
		case "RssAnon":
			anonKB, _ = strconv.ParseUint(strings.TrimSuffix(val, " kB"), 10, 64)
		}
	}
	return name, anonKB, name != ""
}

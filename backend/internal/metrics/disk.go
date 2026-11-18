package metrics

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type DiskUsage struct {
	MountPoint  string  `json:"mount_point"`
	FSType      string  `json:"fs_type"`
	TotalBytes  uint64  `json:"total_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// realFSTypes filtra filesystems virtuales (proc, sysfs, tmpfs, overlay de Docker, etc)
// para reportar solo particiones de disco reales.
var realFSTypes = map[string]bool{
	"ext4": true, "ext3": true, "ext2": true, "xfs": true, "btrfs": true,
	"vfat": true, "exfat": true, "ntfs": true, "ntfs3": true, "f2fs": true,
}

// ReadDiskUsage enumera los mounts reales del host y hace statfs sobre la ruta
// equivalente bajo hostRoot (donde se bind-monteo "/" del host, read-only).
//
// Importante: se lee procPath+"/1/mounts", NO procPath+"/mounts". "/proc/mounts" es un
// symlink a "/proc/self/mounts", y "self" siempre resuelve al namespace de montaje de
// quien esta leyendo -- el propio contenedor del dashboard (sus redes Docker, sus binds
// de /etc/hostname|hosts|resolv.conf) sin importar que /proc este bind-monteado desde
// el host. "/proc/1/mounts" apunta al namespace de montaje del PID 1 especificamente,
// que con pid:host en el compose es el init real del host. Confirmado empiricamente:
// sin este fix, salian los binds internos del propio contenedor como "particiones".
func ReadDiskUsage(procPath, hostRoot string) ([]DiskUsage, error) {
	f, err := os.Open(procPath + "/1/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []DiskUsage
	seen := map[string]bool{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		mountPoint, fsType := fields[1], fields[2]
		if !realFSTypes[fsType] || seen[mountPoint] {
			continue
		}
		seen[mountPoint] = true

		hostPath := filepath.Join(hostRoot, mountPoint)
		var stat syscall.Statfs_t
		if err := syscall.Statfs(hostPath, &stat); err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bfree * uint64(stat.Bsize)
		used := total - free
		var pct float64
		if total > 0 {
			pct = float64(used) / float64(total) * 100
		}

		out = append(out, DiskUsage{
			MountPoint:  mountPoint,
			FSType:      fsType,
			TotalBytes:  total,
			FreeBytes:   free,
			UsedBytes:   used,
			UsedPercent: pct,
		})
	}
	return out, nil
}

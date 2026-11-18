package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func fakeProc(t *testing.T, root, pid, cgroup, status string) {
	t.Helper()
	dir := filepath.Join(root, pid)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "cgroup"), []byte(cgroup), 0o644)
	os.WriteFile(filepath.Join(dir, "status"), []byte(status), 0o644)
}

func TestReadHostProcesses(t *testing.T) {
	root := t.TempDir()
	host := "0::/system.slice/docker.service\n"
	fakeProc(t, root, "10", host, "Name:\tdockerd\nRssAnon:\t   1000 kB\n")
	fakeProc(t, root, "11", "0::/user.slice\n", "Name:\tclaude\nRssAnon:\t 400000 kB\n")
	fakeProc(t, root, "12", "0::/user.slice\n", "Name:\tclaude\nRssAnon:\t  50000 kB\n")
	fakeProc(t, root, "20", "0::/system.slice/docker-abc.scope\n", "Name:\tpython\nRssAnon:\t 999999 kB\n")
	fakeProc(t, root, "2", "0::/\n", "Name:\tkthreadd\n")
	os.MkdirAll(filepath.Join(root, "self"), 0o755)

	got, err := ReadHostProcesses(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("esperaba claude y dockerd (sin contenedor ni kernel), hay %+v", got)
	}
	if got[0].Name != "claude" || got[0].Count != 2 || got[0].MemBytes != 450000*1024 {
		t.Fatalf("claude debe sumar sus 2 procesos e ir primero: %+v", got[0])
	}
}

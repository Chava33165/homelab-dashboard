package dockerapi

import "testing"

func TestStartOrderDependenciesFirst(t *testing.T) {
	c := func(name, service, deps string) ContainerSummary {
		return ContainerSummary{Name: name, Labels: map[string]string{labelService: service, labelDependsOn: deps}}
	}
	// Etiquetas reales del stack de immich en la Pi.
	got := startOrder([]ContainerSummary{
		c("immich_server", "immich-server", "database:service_started:false,redis:service_started:false"),
		c("immich_machine_learning", "immich-machine-learning", ""),
		c("immich_redis", "redis", ""),
		c("immich_postgres", "database", ""),
	})

	pos := map[string]int{}
	for i, m := range got {
		pos[m.Name] = i
	}
	if len(got) != 4 {
		t.Fatalf("esperaba 4 contenedores, hay %d", len(got))
	}
	if pos["immich_server"] < pos["immich_postgres"] || pos["immich_server"] < pos["immich_redis"] {
		t.Fatalf("immich_server debe ir despues de sus dependencias: %v", pos)
	}
}

func TestStartOrderToleratesCycles(t *testing.T) {
	got := startOrder([]ContainerSummary{
		{Name: "a", Labels: map[string]string{labelService: "a", labelDependsOn: "b:service_started:false"}},
		{Name: "b", Labels: map[string]string{labelService: "b", labelDependsOn: "a:service_started:false"}},
	})
	if len(got) != 2 {
		t.Fatalf("un ciclo no debe perder contenedores: %v", got)
	}
}

func TestMarkProtectedWholeSelfProject(t *testing.T) {
	old := selfID
	selfID = "787d4310d162"
	defer func() { selfID = old }()

	list := []ContainerSummary{
		{ID: "787d4310d16298f5", Project: "homelab-dashboard"},
		{ID: "aaaa", Project: "homelab-dashboard"},
		{ID: "bbbb", Project: "immich"},
	}
	markProtected(list)
	if !list[0].Protected || !list[1].Protected || list[2].Protected {
		t.Fatalf("solo el proyecto del dashboard debe quedar protegido: %+v", list)
	}
}

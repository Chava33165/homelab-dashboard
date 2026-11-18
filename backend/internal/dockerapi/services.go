package dockerapi

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/moby/moby/client"
)

var (
	ErrServiceNotFound  = errors.New("servicio no encontrado")
	ErrServiceProtected = errors.New("este servicio corre el propio dashboard y no se puede apagar desde aqui")
)

// SetServicePower prende o apaga todos los contenedores de un proyecto (ver
// ContainerSummary.Project). Respeta el orden de depends_on de compose: al prender,
// dependencias primero (ej. postgres antes que immich_server); al apagar, al reves.
// Va uno por uno a proposito -- en serie es predecible y la Pi no se satura de golpe.
func SetServicePower(ctx context.Context, cli *client.Client, project string, on bool) error {
	all, err := ListContainers(ctx, cli)
	if err != nil {
		return err
	}

	var members []ContainerSummary
	for _, c := range all {
		if c.Project == project {
			members = append(members, c)
		}
	}
	if len(members) == 0 {
		return ErrServiceNotFound
	}
	if members[0].Protected {
		return ErrServiceProtected
	}

	ordered := startOrder(members)
	if !on {
		for i, j := 0, len(ordered)-1; i < j; i, j = i+1, j-1 {
			ordered[i], ordered[j] = ordered[j], ordered[i]
		}
	}

	for _, c := range ordered {
		running := c.State == "running"
		switch {
		case on && !running:
			err = StartContainer(ctx, cli, c.ID)
		case !on && running:
			err = StopContainer(ctx, cli, c.ID)
		default:
			continue
		}
		if err != nil {
			return fmt.Errorf("%s: %w", c.Name, err)
		}
	}
	return nil
}

// startOrder ordena topologicamente por la etiqueta depends_on de compose
// ("database:service_started:false,redis:..."). Un ciclo o una dependencia fuera del
// proyecto no bloquea: lo que no se pueda ordenar se agrega al final por nombre.
func startOrder(members []ContainerSummary) []ContainerSummary {
	sort.Slice(members, func(i, j int) bool { return members[i].Name < members[j].Name })

	byService := map[string][]int{}
	for i, c := range members {
		byService[c.Labels[labelService]] = append(byService[c.Labels[labelService]], i)
	}

	done := make([]bool, len(members))
	visiting := make([]bool, len(members))
	out := make([]ContainerSummary, 0, len(members))

	var visit func(i int)
	visit = func(i int) {
		if done[i] || visiting[i] {
			return
		}
		visiting[i] = true
		for _, dep := range strings.Split(members[i].Labels[labelDependsOn], ",") {
			name, _, _ := strings.Cut(dep, ":")
			if name == "" {
				continue
			}
			for _, j := range byService[name] {
				visit(j)
			}
		}
		visiting[i] = false
		done[i] = true
		out = append(out, members[i])
	}
	for i := range members {
		visit(i)
	}
	return out
}

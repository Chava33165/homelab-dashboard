package dockerapi

import (
	"context"
	"os"
	"strings"

	"github.com/moby/moby/client"
)

type Port struct {
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port,omitempty"`
	Type        string `json:"type"`
}

type ContainerSummary struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Created int64             `json:"created"`
	Ports   []Port            `json:"ports"`
	Labels  map[string]string `json:"-"` // uso interno (resolucion de adapter), no se expone en la API

	// Project agrupa contenedores en "servicios" para la UI: el proyecto de compose, o
	// el propio nombre si el contenedor no viene de compose. Protected marca el
	// proyecto donde corre el propio dashboard -- apagarlo tumbaria la UI que lo apaga.
	Project   string `json:"project"`
	Protected bool   `json:"protected"`
}

const (
	labelProject   = "com.docker.compose.project"
	labelService   = "com.docker.compose.service"
	labelDependsOn = "com.docker.compose.depends_on"
)

func ListContainers(ctx context.Context, cli *client.Client) ([]ContainerSummary, error) {
	res, err := cli.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	out := make([]ContainerSummary, 0, len(res.Items))
	for _, c := range res.Items {
		name := c.ID
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		ports := make([]Port, 0, len(c.Ports))
		for _, p := range c.Ports {
			ports = append(ports, Port{PrivatePort: p.PrivatePort, PublicPort: p.PublicPort, Type: p.Type})
		}

		out = append(out, ContainerSummary{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			State:   string(c.State),
			Status:  c.Status,
			Created: c.Created,
			Ports:   ports,
			Labels:  c.Labels,
			Project: projectOf(name, c.Labels),
		})
	}
	markProtected(out)
	return out, nil
}

func projectOf(name string, labels map[string]string) string {
	if p := labels[labelProject]; p != "" {
		return p
	}
	return name
}

// selfID es el hostname del contenedor del dashboard, que Docker pone por defecto como
// el ID corto del contenedor (compose no le fija "hostname:"). Si no se puede leer,
// queda vacio y simplemente no se marca nada como protegido.
var selfID, _ = os.Hostname()

func markProtected(list []ContainerSummary) {
	if selfID == "" {
		return
	}
	self := ""
	for _, c := range list {
		if strings.HasPrefix(c.ID, selfID) {
			self = c.Project
			break
		}
	}
	for i := range list {
		list[i].Protected = self != "" && list[i].Project == self
	}
}

// IsTTY consulta si un contenedor fue creado con pseudo-TTY -- necesario para saber si
// sus logs vienen como stream unico (tty) o multiplexado stdout/stderr (sin tty).
// Ver logs.go.
func IsTTY(ctx context.Context, cli *client.Client, containerID string) (bool, error) {
	res, err := cli.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return false, err
	}
	if res.Container.Config == nil {
		return false, nil
	}
	return res.Container.Config.Tty, nil
}

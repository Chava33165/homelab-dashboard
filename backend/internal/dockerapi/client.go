// Package dockerapi habla con Docker exclusivamente a traves del docker-socket-proxy
// (nunca monta /var/run/docker.sock directo). El SDK usado es el sucesor oficial del
// antiguo github.com/docker/docker/client -- el proyecto se renombro a moby/moby y el
// modulo Go ahora se declara como github.com/moby/moby/client.
package dockerapi

import "github.com/moby/moby/client"

func NewClient(proxyURL string) (*client.Client, error) {
	return client.New(client.WithHost(proxyURL))
}

package dockerapi

import (
	"context"

	"github.com/moby/moby/client"
)

func StartContainer(ctx context.Context, cli *client.Client, id string) error {
	_, err := cli.ContainerStart(ctx, id, client.ContainerStartOptions{})
	return err
}

func StopContainer(ctx context.Context, cli *client.Client, id string) error {
	_, err := cli.ContainerStop(ctx, id, client.ContainerStopOptions{})
	return err
}

func RestartContainer(ctx context.Context, cli *client.Client, id string) error {
	_, err := cli.ContainerRestart(ctx, id, client.ContainerRestartOptions{})
	return err
}

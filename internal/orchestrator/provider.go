package orchestrator

import (
	"context"
	"errors"
	"fmt"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

func ensureContainerRunning(ctx context.Context, cli *client.Client, image string) (string, error) {
	// warm start
	containerList, err := cli.ContainerList(ctx, client.ContainerListOptions{})
	if err != nil {
		return "", err
	}
	for _, c := range containerList.Items {
		if c.Image == image && c.State == container.StateRunning {
			return c.ID, nil
		}
	}

	// cold start
	containerConfig := &container.Config{
		Image: image,
	}
	hostConfig := &container.HostConfig{
		PublishAllPorts: true,
	}
	createRes, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     containerConfig,
		HostConfig: hostConfig,
	})
	if err != nil {
		return "", fmt.Errorf("create failed: %w", err)
	}
	if _, err = cli.ContainerStart(ctx, createRes.ID, client.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("start failed: %w", err)
	}

	return createRes.ID, nil
}

func discoverContainerURL(ctx context.Context, cli *client.Client, containerID string) (string, error) {
	inspectRes, err := cli.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return "", err
	}
	dockerPort, err := network.ParsePort("8080/tcp")
	if err != nil {
		return "", err
	}
	portBinding := inspectRes.Container.NetworkSettings.Ports[dockerPort]
	if len(portBinding) == 0 {
		return "", errors.New("no port assigned")
	}

	return fmt.Sprintf("http://localhost:%s", portBinding[0].HostPort), nil
}

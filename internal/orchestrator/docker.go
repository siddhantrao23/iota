package orchestrator

import (
	"context"

	"github.com/moby/moby/client"
)

type WorkerResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

func ExecuteCode(cli *client.Client, code string) (string, error) {
	ctx := context.Background()
	imageName := "iota"

	containerID, err := ensureContainerRunning(ctx, cli, imageName)
	if err != nil {
		return "", err
	}

	workerURL, err := discoverContainerURL(ctx, cli, containerID)
	if err != nil {
		return "", err
	}

	return sendCodeToWorker(workerURL, code)
}

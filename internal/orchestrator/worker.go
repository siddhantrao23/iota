package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type Worker struct {
	ID  string
	Url string
}

type WorkerResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

func workerLoop(ctx context.Context, cli *client.Client, id int) {
	defer wg.Done()

	var pendingJob *Job

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		workerInstance, err := createWorker(cli)
		if err != nil {
			fmt.Printf("failed to create worker %d, retrying in 2s...\n", id)
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Printf("✅[Worker %d] Ready at %s\n", id, workerInstance.Url)

		containerHealthy := true

		for containerHealthy {
			if pendingJob == nil {
				select {
				case j := <-JobQueue:
					pendingJob = &j
				case <-ctx.Done():
					containerHealthy = false
				}
			}

			if ctx.Err() != nil {
				break
			}

			out, err := sendCodeToWorker(workerInstance.Url, pendingJob.Code)

			if err != nil && strings.Contains(err.Error(), "worker unreachable") {
				containerHealthy = false
			} else {
				fmt.Printf("📦[Worker %d] Served result to user\n", id)
				pendingJob.ResultChan <- Result{Output: out, Error: err}
				pendingJob = nil
			}
		}

		fmt.Printf("🧹[Worker %d] Removing container\n", id)
		// TODO: fix container cleanup
		_, err = cli.ContainerRemove(context.Background(), workerInstance.ID, client.ContainerRemoveOptions{Force: true})

		if ctx.Err() != nil {
			return
		}
	}
}

func createWorker(cli *client.Client) (*Worker, error) {
	ctx := context.Background()
	image := "iota"

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
		return nil, fmt.Errorf("create failed: %w", err)
	}
	if _, err = cli.ContainerStart(ctx, createRes.ID, client.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("start failed: %w", err)
	}

	inspectRes, err := cli.ContainerInspect(ctx, createRes.ID, client.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}
	dockerPort, err := network.ParsePort("8080/tcp")
	if err != nil {
		return nil, err
	}
	portBinding := inspectRes.Container.NetworkSettings.Ports[dockerPort]
	if len(portBinding) == 0 {
		return nil, errors.New("no port assigned")
	}

	workerUrl := fmt.Sprintf("http://localhost:%s", portBinding[0].HostPort)
	if err := waitForWorker(workerUrl); err != nil {
		_, _ = cli.ContainerRemove(ctx, createRes.ID, client.ContainerRemoveOptions{Force: true})
		return nil, err
	}

	return &Worker{
		ID:  createRes.ID,
		Url: fmt.Sprintf("http://localhost:%s", portBinding[0].HostPort),
	}, nil
}

func waitForWorker(url string) error {
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for worker to boot at %s", url)
		case <-ticker.C:
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				return nil
			}
		}
	}
}

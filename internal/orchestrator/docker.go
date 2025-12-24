package orchestrator

import (
	"bytes"
	"context"
	"strings"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func ExecuteCode(cli *client.Client, code string) (string, error) {
	ctx := context.Background()
	imageName := "iota"

	// create and start container
	containerConfig := &container.Config{
		Image: imageName,
		Cmd:   []string{"python", "-c", code},
	}
	createRes, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{Config: containerConfig})
	if err != nil {
		return "", err
	}
	defer cli.ContainerRemove(ctx, createRes.ID, client.ContainerRemoveOptions{})
	if _, err = cli.ContainerStart(ctx, createRes.ID, client.ContainerStartOptions{}); err != nil {
		return "", err
	}

	waitRes := cli.ContainerWait(ctx, createRes.ID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-waitRes.Error:
		if err != nil {
			return "", err
		}
	case <-waitRes.Result:
	}

	// fetch container logs
	out, err := cli.ContainerLogs(ctx, createRes.ID, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}
	defer out.Close()

	var stdOutBuf, stdErrBuf bytes.Buffer
	if _, err = stdcopy.StdCopy(&stdOutBuf, &stdErrBuf, out); err != nil {
		return "", err
	}
	return strings.TrimSpace(stdOutBuf.String()), nil
}

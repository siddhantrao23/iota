package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func main() {
	ctx := context.Background()
	imageName := "iota"

	apiClient, err := client.New(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	// create and start container
	createRes, err := apiClient.ContainerCreate(ctx, client.ContainerCreateOptions{Image: imageName})
	if err != nil {
		fmt.Printf("Failed to create container %v\n", err)
		return
	}
	fmt.Println("Successfully created container")
	_, err = apiClient.ContainerStart(ctx, createRes.ID, client.ContainerStartOptions{})
	if err != nil {
		fmt.Printf("Failed to create container %v\n", err)
		return
	}
	waitRes := apiClient.ContainerWait(ctx, createRes.ID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-waitRes.Error:
		if err != nil {
			fmt.Printf("Failed to wait for container: %v\n", err)

		}
	case status := <-waitRes.Result:
		fmt.Printf("Container exited with status code: %d\n", status.StatusCode)
	}
	fmt.Println("Successfully started container")

	// fetch container logs
	logs, err := apiClient.ContainerLogs(ctx, createRes.ID, client.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		fmt.Printf("Failed to fetch container output logs %v\n", err)
		return
	}
	io.Copy(os.Stdout, logs)

	// clean up container
	apiClient.ContainerRemove(ctx, createRes.ID, client.ContainerRemoveOptions{})
}

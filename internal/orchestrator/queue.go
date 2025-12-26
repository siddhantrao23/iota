package orchestrator

import (
	"context"
	"sync"

	"github.com/moby/moby/client"
)

type Job struct {
	Code       string
	ResultChan chan Result
}

type Result struct {
	Output string
	Error  error
}

var JobQueue = make(chan Job, 100)
var wg sync.WaitGroup

func StartDispatcher(ctx context.Context, cli *client.Client, workerCount int) {
	for i := range workerCount {
		wg.Add(1)
		go workerLoop(ctx, cli, i)
	}

	go func() {
		<-ctx.Done()
		wg.Wait()
	}()
}

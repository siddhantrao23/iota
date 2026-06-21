package orchestrator

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/moby/moby/client"
)

type Job struct {
	Type       string
	Args       json.RawMessage
	ResultChan chan Result
}

type Result struct {
	Output string
	Error  error
}

var JobQueues = map[string]chan Job{
	"python":     make(chan Job, 20),
	"javascript": make(chan Job, 20),
	"shell":      make(chan Job, 20),
}

var wg sync.WaitGroup

func StartDispatcher(ctx context.Context, cli *client.Client, workers map[string]int) {
	for runtime, count := range workers {
		for i := range count {
			wg.Add(1)
			go workerLoop(ctx, cli, i, runtime)
		}
	}

	go func() {
		<-ctx.Done()
		wg.Wait()
	}()
}

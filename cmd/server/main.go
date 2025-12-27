package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moby/moby/client"
	"github.com/siddhantrao23/iota/internal/api"
	"github.com/siddhantrao23/iota/internal/orchestrator"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cli, err := client.New(client.FromEnv)
	defer cli.Close()
	if err != nil {
		fmt.Printf("Failed to create client%v\n", err)
	}

	orchestrator.StartDispatcher(ctx, cli, 3)

	r := api.SetupRouter()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go srv.ListenAndServe()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	srv.Shutdown(shutdownCtx)
	time.Sleep(2 * time.Second)
}

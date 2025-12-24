package main

import (
	"fmt"

	"github.com/moby/moby/client"
	api "github.com/siddhantrao23/iota/internal/api"
)

func main() {
	cli, err := client.New(client.FromEnv)
	defer cli.Close()
	if err != nil {
		fmt.Printf("Failed to create client%v\n", err)
	}

	router := api.SetupRouter(cli)
	router.Run(":8080")
}

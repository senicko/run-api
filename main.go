package main

import (
	"log"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/senicko/run-api/pkg/api"
	"github.com/senicko/run-api/pkg/pool"
	"github.com/senicko/run-api/pkg/server"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	pool := pool.New(pool.Config{
		Cli:     cli,
		Workers: 10,
	})
	pool.Spawn()

	mux := http.NewServeMux()
	server := server.NewServer(mux, ":8080")

	mux.HandleFunc("/run", api.Method(http.MethodPost, api.Run(pool)))

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}

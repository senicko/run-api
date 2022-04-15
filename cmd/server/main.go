package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/senicko/run-api/pkg/api"
	"github.com/senicko/run-api/pkg/sandbox"
	"github.com/senicko/run-api/pkg/server"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	pool := sandbox.NewPool(cli)
  pool.Start(10)

	mux := http.NewServeMux()
	server := server.NewServer(mux, ":8080")

	mux.HandleFunc("/run", api.Method(http.MethodPost, api.Run(pool)))

  mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "healthy")
  })

	fmt.Println("Starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
package pool

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/senicko/run-api/pkg/sandbox"
)

// Result represents a worker result.
type Result struct {
	Value *sandbox.Response
	Err   error
}

// Request represents a single code execution request.
type Job struct {
	Ctx        context.Context
	RunRequest sandbox.Request
	ResultChan chan<- Result
}

// PoolConfig represents config for a pool.
type Config struct {
	Workers int
	Cli     *client.Client
}

// Pool represents a pool of sandboxes.
type Pool struct {
	cli     *client.Client
	workers int
	JobChan chan Job
}

// New creates a new worker pool.
func New(config Config) *Pool {
	return &Pool{
		cli:     config.Cli,
		workers: config.Workers,
		JobChan: make(chan Job),
	}
}

// Spawn spawns workers.
func (p Pool) Spawn() {
	for i := 0; i < p.workers; i++ {
		go worker(p.cli, p.JobChan)
	}
}

package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	"github.com/senicko/run-api/sandbox"
)

type Response struct {
	Value *sandbox.ExecResponse
	Err   error
}

type Job struct {
	Ctx          context.Context
	ExecRequest  *sandbox.ExecRequest
	ResponseChan chan *Response
}

type Config struct {
	Workers int
	Cli     *client.Client
}

type Pool struct {
	cli          *client.Client
	workersCount int
	jobs         chan *Job
}

// New creates a new worker pool.
func New(config Config) *Pool {
	pool := &Pool{config.Cli, config.Workers, make(chan *Job)}

	for i := 0; i < pool.workersCount; i++ {
		go pool.work()
	}

	return pool
}

// Push pushes a new job to the pool.
func (p Pool) Push(job *Job) {
	p.jobs <- job
}

// work is a pool worker that listens for incoming jobs and handles them.
func (p Pool) work() {
	for job := range p.jobs {
		func() {
			ctx, cancel := context.WithTimeout(job.Ctx, time.Second*5)
			defer cancel()

			s, err := sandbox.CreateSandbox(ctx, p.cli, &job.ExecRequest.Config)
			if err != nil {
				job.ResponseChan <- &Response{Err: fmt.Errorf("pool.work: failed to create the sandbox: %w", err)}
				return
			}
			defer s.Kill()

			response, err := s.Exec(ctx, job.ExecRequest)
      fmt.Println(response, err)

			if err != nil {
				job.ResponseChan <- &Response{Err: fmt.Errorf("pool.work: failed to exec: %w", err)}
				return
			}

			job.ResponseChan <- &Response{Value: response}
		}()
	}
}

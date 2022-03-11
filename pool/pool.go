package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	"github.com/senicko/run-api/sandbox"
)

type Result struct {
	Value *sandbox.Response
	Err   error
}

type Job struct {
	Ctx        context.Context
	RunRequest sandbox.Request
	ResultChan chan *Result
}

type Config struct {
	Workers int
	Cli     *client.Client
}

type Pool struct {
	cli     *client.Client
	workers int
	jobs    chan *Job
}

func New(config Config) *Pool {
	pool := &Pool{config.Cli, config.Workers, make(chan *Job)}

	for i := 0; i < pool.workers; i++ {
		go pool.handle()
	}

	return pool
}

func (p Pool) Push(job *Job) {
	p.jobs <- job
}

func (p Pool) handle() {
	for job := range p.jobs {
		func() {
			ctx, cancel := context.WithTimeout(job.Ctx, time.Second*5)
			defer cancel()

			sID, err := sandbox.PrepareSandbox(ctx, p.cli, job.RunRequest.Language)
			if err != nil {
				job.ResultChan <- &Result{Err: fmt.Errorf("failed to prepare a sandbox: %w", err)}
			}
			defer sandbox.KillSandbox(p.cli, sID)

			response, err := sandbox.Run(ctx, p.cli, sID, job.RunRequest)

			if err != nil {
				job.ResultChan <- &Result{Err: fmt.Errorf("error: failed to run: %w", err)}
				return
			}

			job.ResultChan <- &Result{Value: response}
		}()
	}
}

package sandbox

import (
	"fmt"
	"time"

	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type PoolJob struct {
	Ctx         context.Context
	ExecRequest ExecRequest
	ResultChan  chan PoolResult
}

type PoolResult struct {
	Value *ExecResponse
	Err   error
}

type Pool struct {
	logger  *zap.Logger
	cli     *client.Client
	jobChan chan PoolJob
}

func NewPool(logger *zap.Logger, cli *client.Client) *Pool {
	return &Pool{
		logger:  logger,
		cli:     cli,
		jobChan: make(chan PoolJob),
	}
}

func (p *Pool) Start(count int) {
	for i := 0; i < count; i++ {
		p.spawn()
	}
}

func (p *Pool) Push(job PoolJob) {
	p.jobChan <- job
}

func (p *Pool) spawn() {
	go func() {
		for job := range p.jobChan {
			ctx, cancel := context.WithTimeout(job.Ctx, 5*time.Second)
			defer cancel()

			sandbox, err := CreateSandbox(ctx, p.cli, &job.ExecRequest.Config)
			if err != nil {
				job.ResultChan <- PoolResult{Err: fmt.Errorf("pool.work: failed to create the sandbox: %w", err)}
				return
			}

			result, err := sandbox.Exec(ctx, &job.ExecRequest)
			if err != nil {
				job.ResultChan <- PoolResult{Err: fmt.Errorf("pool.work: failed to exec: %w", err)}
				return
			}

			job.ResultChan <- PoolResult{Value: result}
		}
	}()
}

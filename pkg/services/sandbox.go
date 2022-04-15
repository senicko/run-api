package services

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/senicko/run-api/pkg/sandbox"
	"go.uber.org/zap"
)

type SandboxServiceConfig struct {
	PoolWorkers int // Number of sandbox pool workers
}

type SandboxService struct {
	pool   *sandbox.Pool
	logger *zap.Logger
}

// NewSandboxService creates a new sandbox service.
func NewSandboxService(logger *zap.Logger, config SandboxServiceConfig) (*SandboxService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("services.NewSandboxService: failed to initialize docker client: %w", err)
	}

	pool := sandbox.NewPool(logger, cli)
	pool.Start(config.PoolWorkers)

	return &SandboxService{
		logger: logger,
		pool:   pool,
	}, nil
}

func (s *SandboxService) ProcessRequest(ctx context.Context, execRequest sandbox.ExecRequest) (*sandbox.ExecResponse, error) {
	resultChan := make(chan sandbox.PoolResult)

	s.pool.Push(sandbox.PoolJob{
		Ctx:         ctx,
		ExecRequest: execRequest,
		ResultChan:  resultChan,
	})

	var result sandbox.PoolResult

	// TODO: This select will hang when there won't be any response from sandbox
	// pool. It must be handled in a better way.
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("services.SandboxService.ProcessRequest: context cancelled: %w", ctx.Err())
	case result = <-resultChan:
	}

	if result.Err != nil {
		return nil, fmt.Errorf("services.SandboxService.ProcessRequest: execution failed: %w", result.Err)
	}

	return result.Value, nil
}

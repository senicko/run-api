package pool

import (
	"github.com/docker/docker/client"
	"github.com/senicko/run-api/pkg/sandbox"
)

// worker is a single worker running inside a worker pool.
func worker(cli *client.Client, jobChan <-chan Job) {
	for job := range jobChan {
		ctx := job.Ctx

		if err := ctx.Err(); err != nil {
			job.ResultChan <- Result{
				Err: err,
			}
			return
		}

		response, err := sandbox.Run(ctx, cli, job.RunRequest)
		if err != nil {
			job.ResultChan <- Result{
				Err: err,
			}
			return
		}

		job.ResultChan <- Result{
			Value: response,
		}
	}
}

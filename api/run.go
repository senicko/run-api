package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/senicko/run-api/pool"
	"github.com/senicko/run-api/sandbox"
)

func runInSandbox(ctx context.Context, p *pool.Pool, runRequest sandbox.Request) (*pool.Result, error) {
	var result *pool.Result

	job := &pool.Job{
		Ctx:        ctx,
		RunRequest: runRequest,
		ResultChan: make(chan *pool.Result),
	}
	p.Push(job)

	select {
	case <-ctx.Done():
		return nil, nil
	case result = <-job.ResultChan:
	}

	if result.Err != nil {
		return nil, result.Err
	}

	return result, nil
}

func Run(p *pool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var runRequest sandbox.Request
		if err := json.NewDecoder(r.Body).Decode(&runRequest); err != nil {
			s := http.StatusBadRequest
			http.Error(w, http.StatusText(s), s)
			return
		}

		result, err := runInSandbox(ctx, p, runRequest)
		if err != nil {
      fmt.Println(err)
			s := http.StatusInternalServerError
			http.Error(w, http.StatusText(s), s)
			return
		}

		response, err := json.Marshal(result.Value)
		if err != nil {
			s := http.StatusInternalServerError
			http.Error(w, http.StatusText(s), s)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(response); err != nil {
			s := http.StatusInternalServerError
			http.Error(w, http.StatusText(s), s)
			return
		}
	}
}

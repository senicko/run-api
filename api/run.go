package api

import (
	"encoding/json"
	"net/http"

	"github.com/senicko/run-api/pool"
	"github.com/senicko/run-api/sandbox"
)

// Run is a http request handler for code execution.
func Run(p *pool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var execRequest sandbox.ExecRequest
		if err := json.NewDecoder(r.Body).Decode(&execRequest); err != nil {
			s := http.StatusBadRequest
			http.Error(w, http.StatusText(s), s)
			return
		}

		var result *pool.Response

		job := &pool.Job{
			Ctx:          ctx,
			ExecRequest:  &execRequest,
			ResponseChan: make(chan *pool.Response),
		}
		p.Push(job)

		select {
		case <-ctx.Done():
			return
		case result = <-job.ResponseChan:
		}

		if result.Err != nil {
			s := http.StatusInternalServerError
			http.Error(w, http.StatusText(s), s)
			return
		}

		response, err := json.Marshal(*(result.Value))
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

package api

import (
	"encoding/json"
	"net/http"

	"github.com/senicko/run-api/sandbox"
)

// Run is a http request handler for code execution.
func Run(p *sandbox.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var execRequest sandbox.ExecRequest
		if err := json.NewDecoder(r.Body).Decode(&execRequest); err != nil {
			s := http.StatusBadRequest
			http.Error(w, http.StatusText(s), s)
			return
		}

		resultChan := make(chan sandbox.PoolResult)
		p.Push(sandbox.PoolJob{
			Ctx:         ctx,
			ExecRequest: execRequest,
			ResultChan:  resultChan,
		})

		var result sandbox.PoolResult
		select {
		case <-ctx.Done():
			return
		case result = <-resultChan:
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

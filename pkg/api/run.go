package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/senicko/run-api/pkg/pool"
	"github.com/senicko/run-api/pkg/sandbox"
)

// Run is a request handler that allows to run code.
func Run(p *pool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var runRequest sandbox.Request
		if err := json.NewDecoder(r.Body).Decode(&runRequest); err != nil {
			s := http.StatusBadRequest
			http.Error(w, http.StatusText(s), s)
			return
		}

		resultChan := make(chan pool.Result)
		p.JobChan <- pool.Job{
			Ctx:        ctx,
			RunRequest: runRequest,
			ResultChan: resultChan,
		}

		var result pool.Result
		select {
		case <-ctx.Done():
			fmt.Println("request terminated")
			return
		case result = <-resultChan:
		}
		if result.Err != nil {
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

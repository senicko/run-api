package server

import (
	"net/http"
	"time"
)

// NewServer creates a new http server using passed mux and listening on specified addr.
func NewServer(mux *http.ServeMux, addr string) *http.Server {
	return &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
}

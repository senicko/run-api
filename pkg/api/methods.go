package api

import (
	"net/http"
)

// Methods is a middleware that accepts only specified http methods.
func Method(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if method == r.Method {
			next(w, r)
			return
		}

		s := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(s), s)
	}
}

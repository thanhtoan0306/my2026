package main

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if monitorToken == "" {
			next(w, r)
			return
		}

		if tokenOK(r) {
			next(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized\n"))
	}
}

func tokenOK(r *http.Request) bool {
	if monitorToken == "" {
		return true
	}

	q := r.URL.Query().Get("token")
	if q != "" && subtle.ConstantTimeCompare([]byte(q), []byte(monitorToken)) == 1 {
		return true
	}

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		b := strings.TrimPrefix(auth, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(b), []byte(monitorToken)) == 1 {
			return true
		}
	}

	return false
}


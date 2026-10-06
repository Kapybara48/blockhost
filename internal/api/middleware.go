package api

import (
	"log"
	"net/http"
	"time"
)

// Middleware wraps an http.Handler with CORS and request logging.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)

		// Don't clutter logs with SSE heartbeat streams or static asset requests
		if r.URL.Path != "/api/events/stream" && (r.Method != "GET" || r.URL.Path == "/api/state") {
			log.Printf("[%s] %s - %v", r.Method, r.URL.Path, time.Since(start))
		}
	})
}

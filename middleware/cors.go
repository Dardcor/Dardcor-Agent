package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Silent mode: Only log important actions (like POST) or errors.
		// Ignore noisy asset requests and internal polling.

		isAsset := strings.Contains(r.URL.Path, ".") || strings.HasPrefix(r.URL.Path, "/@") || strings.HasPrefix(r.URL.Path, "/src/")
		isInternal := strings.HasPrefix(r.URL.Path, "/api/ego") || strings.HasPrefix(r.URL.Path, "/api/system") || r.URL.Path == "/health"

		if isAsset || isInternal {
			next.ServeHTTP(w, r)
			return
		}

		// Only log non-GET requests (commands, settings changes) to keep terminal clean
		if r.Method != "GET" {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
			return
		}

		next.ServeHTTP(w, r)
	})
}

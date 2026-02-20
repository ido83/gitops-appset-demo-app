package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// These variables are set at build time via -ldflags in the Dockerfile / CI.
// Keeping them as vars (not consts) is required for -X to work.
var (
	Version   = "0.0.0-dev"
	GitSHA    = "dev"
	BuildTime = "unknown"
)

type Response struct {
	Message   string `json:"message"`
	Version   string `json:"version"`
	GitSHA    string `json:"git_sha"`
	BuildTime string `json:"build_time"`
	TimeUTC   string `json:"time_utc"`
	Hostname  string `json:"hostname"`
}

func main() {
	// Best practice: drive config via env vars for 12-factor compatibility.
	port := getenv("PORT", "8080")

	mux := http.NewServeMux()

	// Readiness/liveness endpoint — kept simple and fast.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// Main endpoint: returns greeting + the version/SHA currently deployed.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host, _ := os.Hostname()
		resp := Response{
			Message:   "Hello from GitOps!",
			Version:   Version,
			GitSHA:    GitSHA,
			BuildTime: BuildTime,
			TimeUTC:   time.Now().UTC().Format(time.RFC3339),
			Hostname:  host,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
	})

	// Timeouts are a production best practice to reduce the blast radius of slow clients.
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("hello-web starting on :%s (version=%s sha=%s buildTime=%s)\n", port, Version, GitSHA, BuildTime)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// loggingMiddleware gives basic request visibility without external deps.
// In real systems you might use structured logging, tracing, etc.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, time.Since(start))
	})
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

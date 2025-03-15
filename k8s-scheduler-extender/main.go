package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
)

// Pod is set up and working
// TODOS: check README

const (
	defaultPort = 8888
)

func main() {
	port := defaultPort
	if p, exists := os.LookupEnv("PORT"); exists {
		if parsedPort, err := strconv.Atoi(p); err == nil {
			port = parsedPort
		}
	}

	// Initialize handlers
	handler := NewExtenderHandler()

	// Set up HTTP server routes
	http.HandleFunc("/filter", handler.Filter)
	http.HandleFunc("/prioritize", handler.Prioritize)

	// Add health check endpoint
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start HTTP server
	listenAddr := fmt.Sprintf(":%d", port)
	klog.Infof("Starting scheduler extender server on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		klog.Fatalf("Failed to start HTTP server: %v", err)
	}
}

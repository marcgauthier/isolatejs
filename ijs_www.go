package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	server = &http.Server{}
)

// Response represents the structure of HTTP response
type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// initializeWebServer sets up and starts the HTTP or HTTPS server
func initializeWebServer(secure bool, certFile, keyFile string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/data", handler(scriptManager))

	addr := fmt.Sprintf("localhost:%d", config.ServerPort)
	server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Check if secure mode is enabled
	if secure {
		// Validate the certificate and key files
		if !fileExists(certFile) || !fileExists(keyFile) {
			logrus.Fatalf("Certificate file (%s) or key file (%s) does not exist", certFile, keyFile)
		}

		go func() {
			logrus.Infof("Starting HTTPS server on %s with cert: %s and key: %s", addr, certFile, keyFile)
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("HTTPS server error: %v", err)
			}
		}()
	} else {
		go func() {
			logrus.Infof("Starting HTTP server on %s", addr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("HTTP server error: %v", err)
			}
		}()
	}
}

// fileExists checks if a file exists and is not a directory
func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// handler processes incoming requests and executes the provided script
func handler(scriptManager *ScriptManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"addr":   r.RemoteAddr,
		}).Info("Received request")

		// Check if the request method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
			logrus.Warn("Request method not allowed")
			return
		}

		// Read and validate request body
		body, err := io.ReadAll(io.LimitReader(r.Body, scriptManager.maxScriptSize))
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			logrus.WithError(err).Error("Failed to read request body")
			return
		}

		// Check if accepting scripts
		if !scriptManager.GetAcceptingScript() {
			http.Error(w, "Currently not accepting script, please wait...", http.StatusMethodNotAllowed)
			logrus.Warn("Rejected script as the system is not accepting scripts")
			return
		}

		logrus.Info("Executing script")
		logrus.Trace(string(body))

		result, execErr := scriptManager.ExecuteScriptWithTimeout(string(body))

		// Prepare response
		response := Response{}
		if execErr != nil {
			handleExecutionError(execErr, w)
			response.Error = execErr.Error()
		} else {
			response.Result = result
			logrus.Info("Script executed successfully, returning result")
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logrus.WithError(err).Error("Failed to encode response")
		}

		logrus.WithFields(logrus.Fields{
			"status": http.StatusOK,
			"took":   time.Since(startTime),
		}).Info("Request processed")
	}
}

// handleExecutionError handles specific script execution errors and sets appropriate HTTP status codes
func handleExecutionError(err error, w http.ResponseWriter) {
	switch err {
	case ErrScriptTooLarge:
		logrus.WithError(err).Warn("Script too large")
		w.WriteHeader(http.StatusBadRequest)
	case ErrNoWorkerAvailable:
		logrus.WithError(err).Warn("No worker available")
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		logrus.WithError(err).Error("Internal server error while processing script")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

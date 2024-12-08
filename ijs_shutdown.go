package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// handleGraceFullShutdown listens for termination signals (SIGINT, SIGTERM),
// gracefully shuts down the server, cancels all running scripts, and performs cleanup.
func handleGraceFullShutdown() {
	// Channel to receive OS signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM) // Listen for interrupt or terminate signals

	// Wait for a signal
	<-stop
	logrus.Info("Shutting down server gracefully...")

	// Create a context with a timeout for shutdown operations
	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeLimit)
	defer cancel()

	// Attempt to gracefully shut down the HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("HTTP server shutdown error")
	}

	// Stop all running scripts using the ScriptManager
	scriptManager.cancelAllScripts()
	logrus.Info("All workers stopped. Exiting after " + config.ShutdownPause.String() + " clean up pause.")

	// Allow time for cleanup operations to complete
	time.Sleep(config.ShutdownPause)
}

// restart performs a graceful shutdown of the server and ScriptManager,
// then restarts the application with the same arguments and environment variables.
// this function is call when the maximum amount of memory is reached for
// a specific amount of time.  The is possibly a script running that is
// grabbing a large amount of memory resources. This could be maliscious code like this:
//
// 		var bigArray = [];
// 		for (var i = 0; i < 1e6; i++) {
// 	 		bigArray.push(new Array(1e3).fill(0)); // Allocate memory in chunks
// 		}

func restart(server *http.Server, scriptManager *ScriptManager) error {
	logrus.Info("Shutting down the server gracefully before restart...")

	// Create a context with a timeout for shutdown operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ensure the server is not nil
	if server == nil {
		logrus.Fatal("Server is nil. Cannot proceed with shutdown.")
	}

	// Attempt to gracefully shut down the HTTP server
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to gracefully shutdown the server: %w", err)
	}

	logrus.Warn("Stopping ScriptManager before restart...")

	// Ensure the ScriptManager is not nil
	if scriptManager == nil {
		logrus.Fatal("ScriptManager is nil. Cannot proceed with shutdown.")
	}

	// Stop all scripts managed by ScriptManager
	scriptManager.Stop()

	// Get the path of the current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable: %w", err)
	}

	// Prepare arguments and environment variables for restarting the application
	args := os.Args
	env := os.Environ()
	logrus.Infof("Restarting application with args: %v", args[1:])

	// Start a new instance of the application with the same arguments
	cmd := exec.Command(exePath, args[1:]...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the new application instance
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to restart application: %w", err)
	}

	logrus.Warn("Exiting current instance to allow restart.")
	os.Exit(0) // Exit the current instance to complete the restart
	return nil
}

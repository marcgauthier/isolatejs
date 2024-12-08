package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

func initializeLogging() {

	// Extract the directory from the log file name
	logDir := filepath.Dir(LogFileName)

	// Create the log directory if it does not exist
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create log directory %s: %v\n", logDir, err)
		os.Exit(1)
	}

	fileLogger := &lumberjack.Logger{
		Filename:   LogFileName,
		MaxSize:    50,
		MaxBackups: 5,
		MaxAge:     90,
		Compress:   true,
	}

	// Set Logrus formatter and output
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	if config.LogOnConsole {
		logrus.SetOutput(io.MultiWriter(fileLogger, os.Stdout))
	} else {
		logrus.SetOutput(fileLogger)
	}

	logrus.SetLevel(VerboseLevel) // Increase verbosity for troubleshooting

	logrus.Info("Logging initialized successfully")

	logrus.Info("Log level is " + VerboseLevel.String())

	// Log the configuration
	logrus.Info(fmt.Sprintf("Logger configuration: Filename=%s, MaxSize=%d MB, MaxBackups=%d, MaxAge=%d days, Compress=%t",
		fileLogger.Filename,
		fileLogger.MaxSize,
		fileLogger.MaxBackups,
		fileLogger.MaxAge,
		fileLogger.Compress,
	))
}

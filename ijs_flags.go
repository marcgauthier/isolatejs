package main

import (
	"flag"
	"fmt"

	"github.com/sirupsen/logrus"
)

var (
	// Global variables for configuration
	VerboseLevel logrus.Level
	ConfigFile   string
	LogFileName  string
)

func init() {
	// Initialize default values
	VerboseLevel = logrus.InfoLevel
	ConfigFile = "./config.yaml"
	LogFileName = "./logs/ijs.log"
}

// ParseFlags parses the command-line flags and sets the global variables
func ParseFlags() {
	// Define the flags
	verboseFlag := flag.String("verbose", "info", "Set the logging level (options: trace, debug, info, warn, error, fatal, panic)")
	configFlag := flag.String("config", "./config.yaml", "Set the configuration file path")
	logFlag := flag.String("log", "./logs/ijs.log", "Set the log file path")

	// Parse the flags
	flag.Parse()

	// Set the global ConfigFile
	ConfigFile = *configFlag

	// Set the log filename
	LogFileName = *logFlag

	// Parse and set the global VerboseLevel
	switch *verboseFlag {
	case "trace":
		VerboseLevel = logrus.TraceLevel
	case "debug":
		VerboseLevel = logrus.DebugLevel
	case "info":
		VerboseLevel = logrus.InfoLevel
	case "warn":
		VerboseLevel = logrus.WarnLevel
	case "error":
		VerboseLevel = logrus.ErrorLevel
	case "fatal":
		VerboseLevel = logrus.FatalLevel
	case "panic":
		VerboseLevel = logrus.PanicLevel
	default:
		fmt.Printf("Invalid verbose level: %s. Defaulting to 'info'.\n", *verboseFlag)
		VerboseLevel = logrus.InfoLevel
	}

	// Apply the log level to logrus
	logrus.SetLevel(VerboseLevel)
}

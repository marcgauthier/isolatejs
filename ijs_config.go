package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	config Config
)

type Config struct {
	MaxMemoryMB       int           `yaml:"max_memory_mb"`
	MaxScriptSize     int64         `yaml:"max_script_size"`
	ServerPort        int           `yaml:"server_port"`
	ScriptTimeout     time.Duration `yaml:"script_timeout"`
	WorkerPoolSize    int           `yaml:"worker_pool_size"`
	LogOnConsole      bool          `yaml:"log_on_console"`
	ShutdownTimeLimit time.Duration `yaml:"shutdown_allow_time"`
	ShutdownPause     time.Duration `yaml:"shutdown_pause_time"`
}

func initializeConfig() {

	var cfg *Config
	var err error

	// config file is set in the flags file
	logrus.Infof("Attempting to load configuration from %s", ConfigFile)
	cfg, err = loadConfig(ConfigFile)
	if err != nil {
		logrus.Fatalf("Error loading %s configuration: %v\n", ConfigFile, err)
	}
	logrus.Infof("Loaded configuration from %s", ConfigFile)

	config = *cfg

	// Validate configuration limits
	if config.MaxMemoryMB < 1 {
		logrus.Fatalf("Invalid memory limit: %d MB, minimum is 1", config.MaxMemoryMB)
	}

	if config.MaxScriptSize < 2 {
		logrus.Fatalf("Invalid script size limit: %d bytes, minimum is 2 {}", config.MaxScriptSize)
	}

	// Log the configuration
	logrus.Info(fmt.Sprintf(
		"Loaded configuration: MaxMemoryMB=%d, MaxScriptSize=%d bytes, ServerPort=%d,ScriptTimeout=%s, GracefulShutdownTimeout=%s, GracefulShutdownPause=%s, WorkerPoolSize=%d, LogOnConsole=%t",
		config.MaxMemoryMB,
		config.MaxScriptSize,
		config.ServerPort,
		config.ScriptTimeout,
		config.ShutdownTimeLimit,
		config.ShutdownPause,
		config.WorkerPoolSize,
		config.LogOnConsole,
	))
}

func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &cfg, nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grafana/sobek"
	"github.com/sirupsen/logrus"
)

// Custom Errors
var (
	ErrScriptTimeout     = errors.New("script execution timed out")
	ErrScriptTooLarge    = errors.New("script size exceeds maximum limit")
	ErrNoWorkerAvailable = errors.New("no worker available to process script")
)

// Global Variables
var (
	scriptManager *ScriptManager
)

// Restricted Globals for the VM Environment
var restrictedGlobals = []string{
	"eval", "process", "child_process", "require", "global", "globalThis",
	"window", "self", "module", "exports", "__dirname", "__filename",
	"XMLHttpRequest", "fetch", "WebSocket", "Object.defineProperty",
	"Object.create", "Proxy", "exec", "execSync", "spawn", "fs",
	"FileSystem", "writeFile", "readFile", "Runtime.getRuntime",
	"setInterval", "setTimeout", "setImmediate", "crypto", "randomBytes",
	"document", "alert", "confirm", "prompt",
}

// ScriptManager handles script execution
type ScriptManager struct {
	sync.RWMutex
	runningScripts  map[string]RunningScriptInfo
	maxScriptSize   int64
	jobQueue        chan ScriptJob
	workerSem       chan struct{}
	cond            *sync.Cond
	scriptCounter   uint64
	acceptingScript int32 // 1 means true, toggled off/on

}

// ScriptJob represents a script job in the queue
type ScriptJob struct {
	Script     string
	ResultChan chan ScriptResult
}

// ScriptResult represents the result of script execution
type ScriptResult struct {
	Result interface{}
	Error  error
}

// RunningScriptInfo stores information about a running script
type RunningScriptInfo struct {
	cancelFunc context.CancelFunc
	vm         *sobek.Runtime
	script     string
}

// Initialize the script manager
func initializeScriptManager() {
	scriptManager = NewScriptManager(config.MaxScriptSize, config.WorkerPoolSize)

	totalCPUs := runtime.NumCPU()
	limitedCPUs := max(1, totalCPUs/2)
	runtime.GOMAXPROCS(limitedCPUs)

	logrus.WithFields(logrus.Fields{
		"Memory Limit (MB)": config.MaxMemoryMB,
		"Max Script Size":   config.MaxScriptSize,
		"CPU Usage":         fmt.Sprintf("%d/%d CPUs", limitedCPUs, totalCPUs),
	}).Info("ScriptManager configuration initialized")
}

// NewScriptManager creates and initializes a new ScriptManager
func NewScriptManager(maxScriptSize int64, workerCount int) *ScriptManager {
	sm := &ScriptManager{
		runningScripts:  make(map[string]RunningScriptInfo),
		maxScriptSize:   maxScriptSize,
		jobQueue:        make(chan ScriptJob, workerCount),
		workerSem:       make(chan struct{}, workerCount),
		acceptingScript: 1,
	}
	sm.cond = sync.NewCond(&sm.RWMutex)

	for i := 0; i < workerCount; i++ {
		go sm.worker()
	}

	go sm.memoryMonitor()
	return sm
}

// Toggle acceptingScript flag
func (sm *ScriptManager) GetAcceptingScript() bool {
	return atomic.LoadInt32(&sm.acceptingScript) == 1
}

func (sm *ScriptManager) setAcceptingScript(v bool) {
	if v {
		atomic.StoreInt32(&sm.acceptingScript, 1)
	} else {
		atomic.StoreInt32(&sm.acceptingScript, 0)
	}
}

// Worker processes jobs from the jobQueue
func (sm *ScriptManager) worker() {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("Worker panic")
		}
		logrus.Info("Worker exiting. Spawning a replacement...")
		go sm.worker() // Maintain pool size
	}()

	for job := range sm.jobQueue {
		ctx, cancel := context.WithTimeout(context.Background(), config.ScriptTimeout)
		sm.workerSem <- struct{}{}

		logrus.WithField("script_length", len(job.Script)).Info("Worker executing script")
		result := sm.executeScript(ctx, job.Script, cancel)
		<-sm.workerSem

		job.ResultChan <- result
		close(job.ResultChan)
	}
}

// Stop cancels all running scripts and resets the job queue
func (sm *ScriptManager) Stop() {
	sm.cancelAllScripts()
}

// ExecuteScript processes a script with a timeout
func (sm *ScriptManager) ExecuteScriptWithTimeout(js string) (interface{}, error) {
	if int64(len(js)) > sm.maxScriptSize {
		logrus.Warn("Script size exceeds maximum limit")
		return nil, ErrScriptTooLarge
	}

	resultChan := make(chan ScriptResult, 1)
	job := ScriptJob{Script: js, ResultChan: resultChan}

	select {
	case sm.jobQueue <- job:
		logrus.WithField("script_length", len(js)).Info("Script queued for execution")
	default:
		logrus.Warn("No available worker for script execution")
		return nil, ErrNoWorkerAvailable
	}

	result := <-resultChan
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Result, nil
}

// Memory Monitor manages resource usage and cancels scripts if memory limits are exceeded
func (sm *ScriptManager) memoryMonitor() {
	var overLimitStart int64
	for {
		time.Sleep(100 * time.Millisecond)
		memStats := &runtime.MemStats{}
		runtime.ReadMemStats(memStats)

		limitBytes := uint64(config.MaxMemoryMB) << 20
		if memStats.Alloc > limitBytes {
			if sm.GetAcceptingScript() {
				sm.setAcceptingScript(false)
				logrus.WithFields(logrus.Fields{
					"usage_mb": memStats.Alloc >> 20,
					"limit_mb": config.MaxMemoryMB,
				}).Warn("Memory usage exceeded limit. Cancelling all scripts...")
				sm.cancelAllScripts()
			}
			overLimitStart = sm.enforceMemoryLimit(overLimitStart)
		} else if overLimitStart != 0 {
			sm.resetMemoryUsage()
			overLimitStart = 0
		}
	}
}

func (sm *ScriptManager) cancelAllScripts() {
	sm.Lock()
	defer sm.Unlock()
	for id, entry := range sm.runningScripts {
		logrus.WithField("script_id", id).Warn("Cancelling script")
		entry.vm.Interrupt("Script cancelled")
		if entry.cancelFunc != nil {
			entry.cancelFunc()
		}
		delete(sm.runningScripts, id)
	}
	logrus.Warn("All scripts cancelled")
}

func (sm *ScriptManager) executeScript(ctx context.Context, js string, cancel context.CancelFunc) ScriptResult {
	vm := sobek.New()

	// Restrict environment
	for _, global := range restrictedGlobals {
		vm.Set(global, nil)
	}

	// Generate a unique script ID
	id := fmt.Sprintf("script-%d", atomic.AddUint64(&sm.scriptCounter, 1))

	// Store the VM and cancelFunc
	sm.Lock()
	sm.runningScripts[id] = RunningScriptInfo{
		cancelFunc: cancel,
		vm:         vm,
		script:     js,
	}
	sm.Unlock()

	resultChan := make(chan ScriptResult, 1)

	go func() {
		defer func() {
			// Once done, remove from runningScripts
			sm.Lock()
			delete(sm.runningScripts, id)
			sm.Unlock()
			vm = nil
		}()

		value, err := vm.RunString(js)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"script_id": id,
				"error":     err,
			}).Error("Script execution failed")
			resultChan <- ScriptResult{Error: fmt.Errorf("script execution failed: %w", err)}
			return
		}
		logrus.WithField("script_id", id).Info("Script completed successfully")
		resultChan <- ScriptResult{Result: value.Export()}
	}()

	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		// Context cancelled: Interrupt the script
		logrus.WithField("script_id", id).Warn("Interrupting script due to context cancellation")
		vm.Interrupt("canceled by context")
		return <-resultChan
	}
}

func (sm *ScriptManager) resetMemoryUsage() {
	logrus.Info("Memory usage back to normal. Resuming script execution in 10 seconds...")
	time.Sleep(10 * time.Second)
	sm.setAcceptingScript(true)
}

func (sm *ScriptManager) enforceMemoryLimit(overLimitStart int64) int64 {
	logrus.Warn("Performing garbage collection due to memory limit")
	runtime.GC()
	debug.FreeOSMemory()
	if overLimitStart == 0 {
		return time.Now().Unix()
	}
	if time.Now().Unix()-overLimitStart > 60 {
		logrus.Error("Memory limit exceeded for over a minute. Restarting...")
		restart(server, scriptManager)
	}
	return overLimitStart
}

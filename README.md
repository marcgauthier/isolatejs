# IsolateJS JavaScript Engine

![IsolateJS Logo](./isolatejs-logo.png "IsolateJS - Secure JavaScript Execution")

This code implements the **IsolateJS JavaScript Engine**, a robust and secure JavaScript execution 
environment designed for tasks such as data analysis, query generation, and dynamic visualization. 
This application is particularly useful for developers working with sensitive data who need a safe 
and efficient way to execute JavaScript code. By leveraging the Goja library, the engine offers 
a controlled runtime environment that emphasizes security, performance, and flexibility.

The core functionality revolves around executing JavaScript scripts in a highly restricted 
and monitored environment. Scripts are subject to stringent safeguards, such as memory and 
execution time limits, to prevent abuse or resource exhaustion. Additionally, the engine 
disallows access to critical system-level features like file operations, network access, and 
system commands, making it ideal for multi-tenant or public-facing applications.

Key features include a configurable system for script execution management, a web server with 
RESTful API endpoints for interaction, and a robust logging mechanism for maintaining 
transparency and debugging. The application also supports dynamic adjustments to CPU allocation 
and resource usage, ensuring stable performance in diverse environments. Graceful shutdown 
mechanisms further enhance the reliability of the system by ensuring all processes terminate 
safely when required.

Overall, this code is tailored for applications requiring secure execution of untrusted or 
user-provided JavaScript, making it a valuable tool for scenarios like data processing 
pipelines, secure script-based automation, and building platforms that allow users to analyze 
and visualize data dynamically within a controlled environment.


## Key Features

### 1. Secure ECMAScript Execution with Goja  
At its core, the IsolateJS JavaScript Engine is built on the **Goja** JavaScript execution engine. By default, Goja provides a restricted execution environment, ensuring untrusted code cannot access external resources unless explicitly configured. This makes it inherently safer for running scripts, especially in multi-tenant or public-facing applications.

### 2. Robust Safety Mechanisms  
The engine addresses potential risks of executing untrusted JavaScript by enforcing strict limits:
- **CPU Exhaustion Protection**:
  - Scripts are restricted to a maximum execution time of **1 second**.
  - CPU usage is capped at **50% of available CPUs**, preventing excessive resource consumption.
- **Memory Exhaustion Protection**:
  - Scripts consuming over **1GB of RAM** are immediately terminated to preserve system stability.

### 3. Extensive Security Restrictions  
By default, the Goja-based runtime disables access to critical system-level functionality, ensuring a secure sandboxed environment for executing user-defined JavaScript:
- **No File System Access**:  
  Untrusted code cannot read or write to the file system. APIs like `fs` and `require('fs')` are unavailable.
- **No Network Access**:  
  Scripts cannot send or receive network requests. APIs like `http`, `fetch`, or `XMLHttpRequest` are excluded.
- **No System Command Execution**:  
  Untrusted code cannot execute shell commands or invoke processes (`child_process`, `exec`).
- **No Native Module Access**:  
  Restricted access to native Node.js modules such as `os`, `path`, or `crypto`.
- **No Global Side Effects**:  
  Scripts cannot alter the global system state outside the Goja runtime.

### 4. Customizable Features  
While the default runtime is highly restricted, developers can extend the engine with custom functionality if needed:
- Add support for **timers** (`setTimeout`, `setInterval`).
- Enable specific APIs selectively to meet application requirements while maintaining security.

### 5. What the Engine Does Not Include  
The IsolateJS JavaScript Engine is not a Node.js runtime or a browser emulator. As such, the following are not included by default:
- **Node.js APIs**: `require`, `process`, `fs`, `http`.
- **Browser APIs**: `window`, `document`, `fetch`, `XMLHttpRequest`.

## Use Cases

- **Data Analysis**: Execute JavaScript to analyze datasets and generate dynamic queries.  
- **Query Generation**: Transform JavaScript logic into database queries for efficient data retrieval.  
- **Visualization**: Generate HTML for charts and graphs to present analysis results dynamically.

## Engine Safeguards Recap

| **Risk**                | **Mitigation**                                                   |
|-------------------------|------------------------------------------------------------------|
| **CPU Exhaustion**      | 1-second limit per script, 50% CPU cap.                         |
| **Memory Exhaustion**   | Scripts exceeding 1GB of RAM are terminated.                    |
| **System Access Risks** | No file, network, or system command access by default.          |

The IsolateJS JavaScript Engine ensures a secure, efficient, and developer-friendly platform for executing and managing JavaScript in sensitive or multi-user environments.


## Recent Updates

### Configuration Enhancements

- Added a new `Config` structure in `IsolateJS_config.go` with fields for:
  - Managing memory limits (`MaxMemoryMB`)
  - Script size (`MaxScriptSize`)
  - Server port (`ServerPort`)
  - Worker pool size (`WorkerPoolSize`)
  - Script timeout (`ScriptTimeout`)

### Logging System
- Introduced a robust logging system in `IsolateJS_logs.go`:
  - `initializeLogging` sets up logging with log rotation, console output, and file-based logs.
  - Log files are maintained in the `logs/` directory, with a maximum size of 50MB per log.

### Script Execution Management
- Enhanced script handling with new structures and functions:
  - `initializeScriptManager` and `NewScriptManager` manage script lifecycle.
  - `GetAcceptingScript` and `SetAcceptingScript` dynamically toggle script acceptance.
  - New structures:
    - `ScriptManager`: Central manager for script execution.
    - `ScriptJob`: Encapsulates details about a script execution job.
    - `RunningScriptInfo`: Tracks ongoing script executions.

### Web Server Enhancements
- A RESTful API was introduced with the `initializeWebServer` function in `IsolateJS_www.go`.
- Standardized API responses using the `Response` structure.

### Graceful Shutdown
- Added `handleGraceFullShutdown` in `IsolateJS_shutdown.go`:
  - Ensures safe and clean termination of services upon receiving shutdown signals.

### Main Function Updates
- Integrated resource management directly into the `main` function:
  - Dynamically allocates up to 50% of available CPUs to prevent resource exhaustion.

## Flags Overview

The `IsolateJS` engine allows configurable runtime behavior using command-line flags. Below are the supported flags:

| **Flag**    | **Description**                                                             | **Default Value**       | **Options**                                    |
|-------------|-----------------------------------------------------------------------------|-------------------------|-----------------------------------------------|
| `-verbose`  | Sets the logging level for the application.                                 | `info`                  | `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic` |
| `-config`   | Specifies the path to the configuration file.                               | `./config.yaml`         | Any valid file path                           |
| `-log`      | Specifies the path to the log file.                                         | `./logs/ijs.log`        | Any valid file path                           |

### Examples:

1. **Set Logging Level to Debug**:
   ```bash
   ./isolatejs -verbose=debug

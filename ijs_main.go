package main

/*

The code implements the **IsolateJS JavaScript Engine**, a robust and secure JavaScript execution
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

When the script is executed the last value of the script is what is returned.

For example:
    var x = 20;
    x;

*/

func main() {

	initializeLogging()

	initializeConfig()

	initializeScriptManager()

	initializeWebServer(false, "", "")

	handleGraceFullShutdown()

}

# concurrent-echo-server

A concurrent HTTP echo server written in Go with graceful shutdown support.

This project demonstrates:

- HTTP server basics using `net/http`
- Concurrency handling under load
- Graceful shutdown using `signal.NotifyContext`
- Request cancellation via `req.Context()`
- Basic health endpoint for shutdown readiness

This project is part of a learning journey toward distributed systems.

---

## Features

- `POST /echo`  
  Echoes the request body back to the client.
- `GET /health`  
  Returns:
  - `200 OK` when running normally
  - `503 Service Unavailable` once shutdown begins

- Graceful shutdown on:
  - `SIGINT` (Ctrl+C)
  - `SIGTERM`

- Simulated request processing delay (2 seconds) to observe:
  - Cancellation behavior
  - Shutdown handling
  - In-flight request completion

---

## Running the Server

### Requirements

- Go 1.20+ (any modern Go version should work)

### Start the server

```bash
go run .
```

### Usage

Run the following in your powershell.

Invoke-RestMethod -Method POST `  -Uri "http://localhost:8000/echo"`
-Body "hello world" `
-ContentType "text/plain"

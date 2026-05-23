# runningMultipleServersInGo

A simple Go HTTP server that runs two servers concurrently on different ports, sharing the same request router and a linked shutdown mechanism.

---

## Overview

This project demonstrates how to run multiple HTTP servers in a single Go program using goroutines, a shared `ServeMux`, and a cancellable `context.Context` to coordinate graceful shutdown across all servers.

---

## File Structure

```
httpserver/
├── main.go     # Entry point — server setup, routing, and lifecycle management
├── go.mod      # Go module definition
└── README.md   # This file
```

---

## How It Works (`main.go`)

### Custom Context Key

```go
type contextKey string
const keyServerAddr contextKey = "serverAddr"
```

A custom `contextKey` type is used to safely store values in a `context.Context` without colliding with keys from other packages. `keyServerAddr` is the key used to inject and retrieve the server's listening address inside each request handler.

---

### Request Handlers

#### `getRoot(w http.ResponseWriter, r *http.Request)`
- Route: `GET /`
- Logs which server received the request (using the address stored in the request context).
- Responds with: `This is my website!`

#### `getHello(w http.ResponseWriter, r *http.Request)`
- Route: `GET /hello`
- Logs which server received the request.
- Responds with: `Hello, HTTP!`

---

### `main()` — Server Lifecycle

#### 1. Router Setup
```go
mux := http.NewServeMux()
mux.HandleFunc("/", getRoot)
mux.HandleFunc("/hello", getHello)
```
A single `ServeMux` routes all incoming requests. Both servers share this same mux.

#### 2. Shared Cancellable Context
```go
ctx, cancelCtx := context.WithCancel(context.Background())
defer cancelCtx()
```
A root context with a cancel function is created. If either server stops for any reason, `cancelCtx()` is called, which signals the rest of the program to shut down.

#### 3. Server Definitions

| Server     | Port  |
|------------|-------|
| `serverOne`| 3000  |
| `serverTwo`| 4444  |

Both servers use `BaseContext` to inject the server's listening address (e.g., `[::]:3000`) into each request's context, so handlers can log which server handled the request.

#### 4. Concurrent Execution
```go
go func() {
    err := serverOne.ListenAndServe()
    ...
    cancelCtx()
}()
```
Each server runs in its own goroutine so both can accept connections simultaneously. If one server fails or closes, it calls `cancelCtx()` to trigger shutdown of the other.

#### 5. Blocking Until Shutdown
```go
<-ctx.Done()
```
`main` blocks here until the shared context is cancelled.

#### 6. Graceful Shutdown
```go
serverOne.Shutdown(context.Background())
serverTwo.Shutdown(context.Background())
```
Both servers are shut down gracefully, allowing any in-flight requests to complete before the program exits.

---

## Running the Server

```bash
go run main.go
```

The server will start listening on two ports:

| URL                          | Handler    |
|------------------------------|------------|
| http://localhost:3000/       | `getRoot`  |
| http://localhost:3000/hello  | `getHello` |
| http://localhost:4444/       | `getRoot`  |
| http://localhost:4444/hello  | `getHello` |

---

## Example Output

```
[::]:3000: got / request
[::]:4444: got /hello request
```

---

## Requirements

- Go 1.24.2 or later

---

## Module

```
github.com/vision7r/httpserver
```

# runningMultipleServersInGo

A simple Go HTTP server that runs two servers concurrently on different ports, sharing the same request router and a linked shutdown mechanism.

---

## Overview

This project demonstrates how to run multiple HTTP servers in a single Go program using goroutines, a shared `ServeMux`, and a cancellable `context.Context` to coordinate graceful shutdown across all servers.


### Custom Context Key

```go
type contextKey string
const keyServerAddr contextKey = "serverAddr"
```

A custom `contextKey` type is used to safely store values in a `context.Context` without colliding with keys from other packages. `keyServerAddr` is the key used to inject and retrieve the server's listening address inside each request handler.

---


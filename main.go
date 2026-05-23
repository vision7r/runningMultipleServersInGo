package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

// contextKey is a custom type for context keys to avoid collisions with other packages.
type contextKey string

// keyServerAddr is used to store and retrieve the server's listening address from a request context.
const keyServerAddr contextKey = "serverAddr"

// getRoot handles GET requests to "/". It logs which server received the request
// and responds with a simple message.
func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s: got / request\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "This is my website!\n")
}

// getHello handles GET requests to "/hello". It logs which server received the request
// and responds with a greeting.
func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("%s: got /hello request\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
	// mux routes incoming requests to the correct handler based on the URL path.
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/hello", getHello)

	// ctx is a shared cancellable context. When either server stops, cancelCtx is called
	// to signal the other server (and main) to shut down as well.
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// serverOne listens on port 3000. BaseContext injects the server's address into
	// each request's context so handlers can log which server handled the request.
	serverOne := &http.Server{
		Addr:    ":3000",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return context.WithValue(ctx, keyServerAddr, l.Addr().String())
		},
	}

	// serverTwo listens on port 4444, sharing the same mux and context setup as serverOne.
	serverTwo := &http.Server{
		Addr:    ":4444",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return context.WithValue(ctx, keyServerAddr, l.Addr().String())
		},
	}

	// Run each server in its own goroutine so both can accept connections concurrently.
	// If a server stops for any reason, cancelCtx triggers a shutdown of everything.
	go func() {
		err := serverOne.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server one closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server one: %s\n", err)
		}
		cancelCtx()
	}()

	go func() {
		err := serverTwo.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server two closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server two: %s\n", err)
		}
		cancelCtx()
	}()

	// Block here until the context is cancelled (i.e., one of the servers stops).
	<-ctx.Done()

	// Gracefully shut down both servers, allowing in-flight requests to finish.
	if err := serverOne.Shutdown(context.Background()); err != nil {
		fmt.Printf("error shutting down server one: %s\n", err)
	}
	if err := serverTwo.Shutdown(context.Background()); err != nil {
		fmt.Printf("error shutting down server two: %s\n", err)
	}
}

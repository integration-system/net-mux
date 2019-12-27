
# Example
```go
package main

import (
	"context"
	mux "github.com/integration-system/net-mux"
	"log"
	"net"
	"net/http"
)

func main() {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("create listener: %v", err)
	}

	muxer := mux.New(tcpListener)
	httpListener := muxer.Match(mux.HTTP1())
	customProtocolListener := muxer.Match(mux.Any())

	go func() {
		if err := muxer.Serve(); err != nil {
			log.Printf("serve mux: %v\n", err)
		}
	}()

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// ...
	})
	httpServer := &http.Server{Handler: httpMux}
	go func() {
		if err := httpServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
			log.Printf("http server closed: %v", err)
		}
	}()

	// Use customProtocolListener

	// ...
	
	_ = httpServer.Shutdown(context.Background())
	if err := muxer.Close(); err != nil {
		log.Fatalf("close listener: %v", err)
	}
}
```
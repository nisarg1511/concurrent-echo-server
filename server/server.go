package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	_shutdownPeriod      = 15 * time.Second
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

var isShuttingDown atomic.Bool

func Serve() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//Endpoints
	http.HandleFunc("/echo", handleEchoRequest)

	http.HandleFunc("/health", handleHealthCheckRequuest)
	//Server configurations

	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: ":8000",
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Println("Server starting on :8000")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-rootCtx.Done()

	isShuttingDown.Store(true)
	time.Sleep(_readinessDrainDelay)
	log.Println("Readiness check propogated now waiting for ongoing requests to finish.s")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)

	defer cancel()
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		log.Printf("Shutdown error: %v; forcing close", err)
		_ = server.Close()
	}
	stopOngoingGracefully()
	if err != nil {
		log.Println("Failed to wait for ongoing requests to finish, waiting for forced cancellation.")
		time.Sleep(_shutdownHardPeriod)
	}
	log.Println("Server shut down gracefully.")
}

func handleHealthCheckRequuest(w http.ResponseWriter, r *http.Request) {
	if isShuttingDown.Load() {
		http.Error(w, "Shtting down", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "OK")
}

func handleEchoRequest(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	if req.Method != http.MethodPost {
		http.Error(res, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	select {
	case <-time.After(2 * time.Second):
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("%v\n", err)
			res.Write([]byte("Error occured while processing the request!"))
			return
		}
		fmt.Fprintf(res, "%s", string(body))

	case <-req.Context().Done():
		http.Error(res, "Request canecelled", http.StatusRequestTimeout)
	}

}

func handleReadiness(w http.ResponseWriter, _ *http.Request) {
	if isShuttingDown.Load() {
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "ok")
}

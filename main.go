package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ErrorResponse model
type ErrorResponse struct {
	Message string `json:"error"`
}

func main() {

	// Initalize PasswordHash
	passwordHash := PasswordHash{
		passwords:      make(map[int]string),
		id:             0,
		totalTimeMicro: 0,
		sleeper:        &defaultSleeper{},
		shutdown:       make(chan os.Signal, 1),
		waitGroup:      new(sync.WaitGroup),
	}

	// Notify the shutdown channel on os signals
	signal.Notify(passwordHash.shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{Addr: ":8080"}

	http.Handle("/hash", hashHandler{&passwordHash})
	http.Handle("/hash/", hashHandler{&passwordHash})

	http.Handle("/stats", statsHandler{&passwordHash})
	http.Handle("/stats/", statsHandler{&passwordHash})

	http.Handle("/shutdown", shutdownHandler{&passwordHash})
	http.Handle("/shutdown/", shutdownHandler{&passwordHash})

	// Listen and Serve in a goroutine so main can handle listening on the shutdown channel
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	fmt.Println("Server started")

	// Listen on shutdown channel
	<-passwordHash.shutdown
	fmt.Println("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		passwordHash.waitGroup.Wait()
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Print("Server shutdown")
}

package main

import (
	"fmt"
	"net/http"
	"os"
)

// Handler for the shutdown endpoint
type shutdownHandler struct {
	passwordHash *PasswordHash
}

// Handler for the shutdown endpoint
func (shutdownh shutdownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Shutdown called")
	// Send interrupt signal to shutdown channel to trigger graceful shutdown
	shutdownh.passwordHash.shutdown <- os.Interrupt
}

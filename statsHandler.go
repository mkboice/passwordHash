package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type stats struct {
	Total   int
	Average int
}

// Handler for the stats endpoints
type statsHandler struct {
	passwordHash *PasswordHash
}

// Handler for the stats endpoints
func (sh statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		numRequests, averageTimePerRequest := sh.passwordHash.getRequestStats()
		realStats := stats{Total: numRequests, Average: averageTimePerRequest}

		jsonResult, ok := json.Marshal(&realStats)
		if ok != nil {
			log.Fatal(ok)
		}
		log.Printf("stats: %v", string(jsonResult))
		fmt.Fprintf(w, string(jsonResult))
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

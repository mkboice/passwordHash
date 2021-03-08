package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ErrorResponse model
type ErrorResponse struct {
	Message string `json:"error"`
}

// Define the sleeper interface that has the Sleep method
// This allows the unit tests to replace this functionality to avoid unnecessary sleeps
type sleeper interface {
	Sleep()
}

// DefaultSleeper struct
type defaultSleeper struct{}

// Define the Sleep functionality for the defaultSleeper
func (d *defaultSleeper) Sleep() {
	time.Sleep(5 * time.Second)
}

// PasswordHash struct to keep track of hash and id data
// Passwords and ID are stored in memory but could easily be replaced with a database connection for persistent data
type PasswordHash struct {
	passwords      map[int]string  // Map of id and hash of password
	id             int             // ID counter
	totalTimeMicro int64           // Total time all requests have taken in microseconds
	sleeper        sleeper         // Sleeper interface so unit tests can replace it
	shutdown       chan os.Signal  // Shutdown channel for graceful shutdown
	waitGroup      *sync.WaitGroup // Waitgroup for making sure goroutines finish for graceful shutdown
}

// Increment the total time in microseconds based on the elapsed time
func (ph *PasswordHash) incrementTime(start time.Time) {
	elapsed := time.Since(start)
	ph.totalTimeMicro = ph.totalTimeMicro + elapsed.Microseconds()
}

// Increment the id
func (ph *PasswordHash) incrementID() {
	ph.id = ph.id + 1
}

// TODO: Understand pointer part better
// Comptue the hash given the id and password string and save it in the passwords map
func (ph *PasswordHash) computeHash(id int, password string) {
	defer ph.waitGroup.Done()
	fmt.Printf("COMPTUE called id: %s password: %s\n", strconv.Itoa(id), password)
	ph.sleeper.Sleep()
	hash := sha512.Sum512([]byte(password))
	sha512Hash := base64.StdEncoding.EncodeToString(hash[:])
	fmt.Println("id: " + strconv.Itoa(id) + " hash: " + sha512Hash)
	ph.passwords[id] = sha512Hash
}

// Get the hash by id
func (ph *PasswordHash) getHashByID(id int) (string, bool) {
	fmt.Printf("GET id: %v hash: %s\n", id, ph.passwords[id])
	hash, found := ph.passwords[id]

	return hash, found
}

// Get the number of requests and average time per request in microseconds
func (ph *PasswordHash) getRequestStats() (int, int) {
	var numRequests int
	var averageTimeMicroseconds int
	if ph.id <= 0 || ph.totalTimeMicro <= 0 {
		numRequests = 0
		averageTimeMicroseconds = 0
	} else {
		numRequests = ph.id
		averageTimeMicroseconds = int(ph.totalTimeMicro) / ph.id
	}
	return numRequests, averageTimeMicroseconds
}

// Handler for the hash endpoints
type hashHandler struct {
	passwordHash *PasswordHash
}

// get the id from the url path
func (hh hashHandler) getIDFromPath(url *url.URL) (int, error) {
	fmt.Println(url)
	path := url.Path
	splitPath := strings.Split(path, "/")
	// Get the last element in the path
	getIDString := splitPath[len(splitPath)-1]
	fmt.Println("id: " + getIDString)
	getID, ok := strconv.Atoi(getIDString)
	if ok != nil {
		return 0, fmt.Errorf("Invalid ID %v", getIDString)
	}

	return getID, ok
}

// Handler for the hash endpoints
func (hh hashHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		getID, ok := hh.getIDFromPath(r.URL)
		if ok != nil {
			errorResponse := ErrorResponse{Message: ok.Error()}
			jsonResponse, ok := json.Marshal(&errorResponse)
			if ok != nil {
				log.Fatal(ok)
			}
			fmt.Println(string(jsonResponse))
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, string(jsonResponse))
			return
		}

		hash, found := hh.passwordHash.getHashByID(getID)
		if !found {
			errorResponse := ErrorResponse{Message: fmt.Sprintf("ID %v not found", getID)}
			jsonResponse, ok := json.Marshal(&errorResponse)
			if ok != nil {
				log.Fatal(ok)
			}
			fmt.Println(string(jsonResponse))
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, string(jsonResponse))
			return
		}

		fmt.Fprintf(w, "%s", hash)
		return

	case http.MethodPost:
		defer hh.passwordHash.incrementTime(time.Now())
		hh.passwordHash.incrementID()
		password := r.PostFormValue("password")
		if password == "" {
			errorResponse := ErrorResponse{Message: "Bad request: empty password"}
			jsonResponse, ok := json.Marshal(&errorResponse)
			if ok != nil {
				log.Fatal(ok)
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, string(jsonResponse))
			return
		}

		fmt.Println("password: " + password)

		// call go routine
		hh.passwordHash.waitGroup.Add(1)
		go hh.passwordHash.computeHash(hh.passwordHash.id, password)

		fmt.Fprintf(w, strconv.Itoa(hh.passwordHash.id))
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

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
		fmt.Fprintf(w, string(jsonResult))
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

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

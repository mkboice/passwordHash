package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ErrorResponse model
type ErrorResponse struct {
	Message string `json:"error"`
}

type Sleeper interface {
	Sleep()
}

type DefaultSleeper struct{}

func (d *DefaultSleeper) Sleep() {
	time.Sleep(5 * time.Second)
}

type PasswordHash struct {
	Passwords      map[int]string
	Id             int
	TotalTimeMicro int64
	Sleeper        Sleeper
	shutdown       chan os.Signal
}

type passwordHashHandler struct {
	passwordHash *PasswordHash
}

type shutdownHandler struct {
	passwordHash *PasswordHash
}

func (phh passwordHashHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println(r.URL)
		path := r.URL.Path
		splitPath := strings.Split(path, "/") // TODO: Add verification logic
		// Get the last element in the path
		getID := splitPath[len(splitPath)-1]
		fmt.Println("id: " + getID)
		getIDInt, ok := strconv.Atoi(getID)
		if ok != nil {
			errorResponse := ErrorResponse{Message: "Invalid ID " + getID}
			jsonResponse, ok := json.Marshal(&errorResponse)
			if ok != nil {
				log.Fatal(ok)
			}
			fmt.Println(string(jsonResponse))
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, string(jsonResponse))
			return
		}

		fmt.Println("GET id: " + getID + " hash: " + phh.passwordHash.Passwords[getIDInt])
		hash, found := phh.passwordHash.Passwords[getIDInt]
		if !found {
			errorResponse := ErrorResponse{Message: "ID " + getID + " not found"}
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
		defer phh.incrementTime(time.Now()) // Making the assumption that average time is asking about the actual request, not hashing it
		phh.passwordHash.Id = phh.passwordHash.Id + 1
		password := r.PostFormValue("password")
		if password == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Bad request: password not in form")
			return
		}

		fmt.Println("password: " + password)
		// call go routine
		go phh.computeHash(password, phh.passwordHash.Id)
		fmt.Fprintf(w, strconv.Itoa(phh.passwordHash.Id))
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (phh *passwordHashHandler) incrementTime(start time.Time) {
	elapsed := time.Since(start)
	phh.passwordHash.TotalTimeMicro = phh.passwordHash.TotalTimeMicro + elapsed.Microseconds()
}

// TODO: Understand pointer part better
func (phh *passwordHashHandler) computeHash(password string, id int) {
	fmt.Printf("COMPTUE called id: %s password: %s\n", strconv.Itoa(id), password)
	phh.passwordHash.Sleeper.Sleep()
	hash := sha512.Sum512([]byte(password))
	sha512Hash := base64.StdEncoding.EncodeToString(hash[:])
	fmt.Println("id: " + strconv.Itoa(id) + " hash: " + sha512Hash)
	phh.passwordHash.Passwords[id] = sha512Hash
}

type stats struct {
	Total   int
	Average int
}

type statsHandler struct {
	passwordHash *PasswordHash
}

func (sh statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var realStats stats
		if sh.passwordHash.Id <= 0 || sh.passwordHash.TotalTimeMicro <= 0 {
			realStats = stats{Total: 0, Average: 0}
		} else {
			realStats = stats{Total: sh.passwordHash.Id, Average: int(sh.passwordHash.TotalTimeMicro) / sh.passwordHash.Id}
		}

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

func (shutdownh shutdownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Shutdown called")
	shutdownh.passwordHash.shutdown <- os.Interrupt
}

func main() {

	passwordHash := PasswordHash{Passwords: make(map[int]string), Id: 0, TotalTimeMicro: 0, Sleeper: &DefaultSleeper{}, shutdown: make(chan os.Signal, 1)}
	signal.Notify(passwordHash.shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{Addr: ":8080"}
	http.Handle("/hash/", passwordHashHandler{&passwordHash})
	http.Handle("/hash", passwordHashHandler{&passwordHash})

	http.Handle("/stats", statsHandler{&passwordHash})
	http.Handle("/shutdown", shutdownHandler{&passwordHash})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	fmt.Println("Server started")

	<-passwordHash.shutdown
	fmt.Println("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// Do extra handling here
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Print("Server shutdown")
}

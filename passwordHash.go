package main

import (
	"crypto/sha512"
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

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
	ph.id++
}

// Update the id and start the goroutine to compute the hash
func (ph *PasswordHash) computeHash(password string) {
	ph.incrementID()
	ph.waitGroup.Add(1)
	go ph.doComputeHash(ph.id, password)
}

// Comptue the hash given the id and password string and save it in the passwords map
func (ph *PasswordHash) doComputeHash(id int, password string) {
	defer ph.waitGroup.Done()
	log.Printf("Creating hash for id: %s password: %s\n", strconv.Itoa(id), password)
	ph.sleeper.Sleep()
	hash := sha512.Sum512([]byte(password))
	sha512Hash := base64.StdEncoding.EncodeToString(hash[:])
	log.Println("id: " + strconv.Itoa(id) + " hash: " + sha512Hash)
	ph.passwords[id] = sha512Hash
}

// Get the hash by id
func (ph *PasswordHash) getHashByID(id int) (string, bool) {
	log.Printf("GET id: %v hash: %s\n", id, ph.passwords[id])
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

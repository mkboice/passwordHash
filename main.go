package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ErrorResponse model
type ErrorResponse struct {
	Message string `json:"error"`
}

// type PasswordHash struct {
// 	Passwords      map[int]string
// 	Id             int
// 	TotalTimeMicro int64
// }

var id int = 0
var totalTimeMicro int64 = 0

// Passwords map of id and hash
var Passwords map[int]string = make(map[int]string)

func averageTime(start time.Time) {
	elapsed := time.Since(start)
	totalTimeMicro = totalTimeMicro + elapsed.Microseconds()
}

func computeHash(password string, id int) {
	fmt.Printf("COMPTUE called id: %s password: %s\n", strconv.Itoa(id), password)
	time.Sleep(5 * time.Second)
	hash := sha512.Sum512([]byte(password))
	sha512Hash := base64.StdEncoding.EncodeToString(hash[:])
	fmt.Println("id: " + strconv.Itoa(id) + " hash: " + sha512Hash)
	Passwords[id] = sha512Hash
}

// HashHandler blah
func HashHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		defer averageTime(time.Now()) // Making the assumption that average time is asking about the actual request, not hashing it
		id = id + 1
		password := r.PostFormValue("password")
		if password == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Bad request: password not in form")
			return
		}

		fmt.Println("password: " + password)
		// call go routine
		go computeHash(password, id)
		fmt.Fprintf(w, strconv.Itoa(id))
		return

	} else if r.Method == http.MethodGet {
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

		fmt.Println("GET id: " + getID + " hash: " + Passwords[getIDInt])
		hash, found := Passwords[getIDInt]
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
	}
}

type stats struct {
	Total   int
	Average int
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	realStats := stats{Total: id, Average: int(totalTimeMicro) / id}
	jsonResult, ok := json.Marshal(&realStats)
	if ok != nil {
		log.Fatal(ok)
	}
	fmt.Fprintf(w, string(jsonResult))
}

func main() {
	// passwordHash = PasswordHash{Passwords: make(map[int]string), Id: 0, TotalTimeMicro: 0}

	http.HandleFunc("/hash/", HashHandler) // TODO: Does not handle hash without ending '/' should I add another route to handle that case?
	http.HandleFunc("/hash", HashHandler)

	http.HandleFunc("/stats", statsHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

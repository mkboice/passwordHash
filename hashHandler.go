package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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
		hh.passwordHash.computeHash(password)

		fmt.Fprintf(w, strconv.Itoa(hh.passwordHash.id))
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

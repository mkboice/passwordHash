package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

type SpySleeper struct {
	Calls int
}

func (s *SpySleeper) Sleep() {
	s.Calls++
}

func TestHashHandler(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		input      *PasswordHash
		path       string
		password   string
		expected   string
		statusCode int
	}{
		{
			name:   "Get hash id 4 which does not exist",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/hash/4",
			password:   "",
			expected:   `{"error":"ID 4 not found"}`,
			statusCode: http.StatusNotFound,
		},
		{
			name:   "Get invalid hash id",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/hash/4abc",
			password:   "",
			expected:   `{"error":"Invalid ID 4abc"}`,
			statusCode: http.StatusNotFound,
		},
		{
			name:   "Hash angryMonkey",
			method: http.MethodPost,
			input: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/hash",
			password:   "angryMonkey",
			expected:   "1",
			statusCode: http.StatusOK,
		},
		{
			name:   "Hash 123",
			method: http.MethodPost,
			input: &PasswordHash{
				passwords:      map[int]string{1: "hashstring"},
				id:             1,
				totalTimeMicro: 123,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/hash",
			password:   "123",
			expected:   "2",
			statusCode: http.StatusOK,
		},
		{
			name:   "Hash something",
			method: http.MethodPost,
			input: &PasswordHash{
				passwords:      map[int]string{1: "hashstring", 2: "anotherhash"},
				id:             2,
				totalTimeMicro: 145,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/hash/",
			password:   "something",
			expected:   "3",
			statusCode: http.StatusOK,
		},
		{
			name:   "Hash empty password",
			method: http.MethodPost,
			input: &PasswordHash{
				passwords:      map[int]string{1: "hashstring", 2: "anotherhash"},
				id:             2,
				totalTimeMicro: 145,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/hash/",
			password:   "",
			expected:   `{"error":"Bad request: empty password"}`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "Get hash id 1",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{1: "hashstring", 2: "anotherhash", 3: "yetanotherone"},
				id:             3,
				totalTimeMicro: 137,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:     "/hash/1",
			password: "",
			// expected:   "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==",
			expected:   "hashstring",
			statusCode: http.StatusOK,
		},
		{
			name:   "Get hash id 2",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{1: "hashstring", 2: "anotherhash", 3: "yetanotherone"},
				id:             3,
				totalTimeMicro: 137,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:     "/hash/2",
			password: "",
			// expected:   "PJkJr+wlNU1VHa4hWQuybjjVPyFzuNPcPu5MBH56scHri4UQPjvnumE7MbtcnDYhTcnxSkL9ei/bhIVrylxEwg==",
			expected:   "anotherhash",
			statusCode: http.StatusOK,
		},
		{
			name:   "Get hash id 3",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{1: "hashstring", 2: "anotherhash", 3: "yetanotherone"},
				id:             3,
				totalTimeMicro: 137,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:     "/hash/3",
			password: "",
			// expected:   "mD1D3f9tqQ9qXTthckRqH/4ii4A/5k/dXc+rVkYHioloUf6C9iPJ1uVlSz0vNjoE7BfPtitgdDepx8Ey1RHlIg==",
			expected:   "yetanotherone",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			var err error
			if tc.method == http.MethodGet {
				req, err = http.NewRequest(tc.method, "http://localhost:8080"+tc.path, nil)
			}

			if tc.method == http.MethodPost {
				fmt.Println("In post setup")
				formData := url.Values{}
				formData.Add("password", tc.password)
				req, err = http.NewRequest(tc.method, "http://localhost:8080"+tc.path, strings.NewReader(formData.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			}

			if err != nil {
				t.Error(err)
			}
			resRecorder := httptest.NewRecorder()

			hashHandler{tc.input}.ServeHTTP(resRecorder, req)

			// Check status code
			status := resRecorder.Code
			if status != tc.statusCode {
				t.Errorf("Expected %v but instead got %v", tc.statusCode, status)
			}

			// Check body
			body := strings.TrimSpace(resRecorder.Body.String())
			if body != tc.expected {
				t.Errorf("Expected %v but instead got %v", tc.expected, body)
			}

		})
	}
}

func TestStatsHandler(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		input      *PasswordHash
		path       string
		password   string
		expected   string
		statusCode int
	}{
		{
			name:   "Get stats after no passwords were hashed",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/stats",
			password:   "",
			expected:   `{"Total":0,"Average":0}`,
			statusCode: http.StatusOK,
		},
		{
			name:   "Get stats after a few passwords were hashed",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{},
				id:             3,
				totalTimeMicro: 789,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/stats",
			password:   "",
			expected:   `{"Total":3,"Average":263}`,
			statusCode: http.StatusOK,
		},
		{
			name:   "Post stats not allowed",
			method: http.MethodPost,
			input: &PasswordHash{
				passwords:      map[int]string{},
				id:             3,
				totalTimeMicro: 789,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/stats",
			password:   "",
			expected:   "Method not allowed",
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			req, err := http.NewRequest(tc.method, "http://localhost:8080"+tc.path, nil)

			if err != nil {
				t.Error(err)
			}
			resRecorder := httptest.NewRecorder()

			statsHandler{tc.input}.ServeHTTP(resRecorder, req)

			// Check status code
			status := resRecorder.Code
			if status != tc.statusCode {
				t.Errorf("Expected %v but instead got %v", tc.statusCode, status)
			}

			// Check body
			body := strings.TrimSpace(resRecorder.Body.String())
			if body != tc.expected {
				t.Errorf(`Expected "%v" but instead got "%v"`, tc.expected, body)
			}

		})
	}
}

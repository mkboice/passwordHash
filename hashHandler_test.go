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
			name:   "Get hash id 1",
			method: http.MethodGet,
			input: &PasswordHash{
				passwords:      map[int]string{1: "hashstring", 2: "anotherhash", 3: "yetanotherone"},
				id:             3,
				totalTimeMicro: 137,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			path:       "/hash/1",
			password:   "",
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
			path:       "/hash/2",
			password:   "",
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
			path:       "/hash/3",
			password:   "",
			expected:   "yetanotherone",
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

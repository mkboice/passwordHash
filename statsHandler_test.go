package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

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

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestHashHandler(t *testing.T) {
	cases := []struct {
		name       string
		method     string
		path       string
		password   string
		expected   string
		statusCode int
	}{
		{
			name:       "Get hash id 4 which does not exist",
			method:     http.MethodGet,
			path:       "/hash/4",
			password:   "",
			expected:   `{"error":"ID 4 not found"}`,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Get invalid hash id",
			method:     http.MethodGet,
			path:       "/hash/4abc",
			password:   "",
			expected:   `{"error":"Invalid ID 4abc"}`,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Hash angryMonkey",
			method:     http.MethodPost,
			path:       "/hash",
			password:   "angryMonkey",
			expected:   "1",
			statusCode: http.StatusOK,
		},
		{
			name:       "Hash 123",
			method:     http.MethodPost,
			path:       "/hash",
			password:   "123",
			expected:   "2",
			statusCode: http.StatusOK,
		},
		{
			name:       "Hash something",
			method:     http.MethodPost,
			path:       "/hash/",
			password:   "something",
			expected:   "3",
			statusCode: http.StatusOK,
		},
		{
			name:       "Get hash id 1",
			method:     http.MethodGet,
			path:       "/hash/1",
			password:   "",
			expected:   "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==",
			statusCode: http.StatusOK,
		},
		{
			name:       "Get hash id 2",
			method:     http.MethodGet,
			path:       "/hash/2",
			password:   "",
			expected:   "PJkJr+wlNU1VHa4hWQuybjjVPyFzuNPcPu5MBH56scHri4UQPjvnumE7MbtcnDYhTcnxSkL9ei/bhIVrylxEwg==",
			statusCode: http.StatusOK,
		},
		{
			name:       "Get hash id 3",
			method:     http.MethodGet,
			path:       "/hash/3",
			password:   "",
			expected:   "mD1D3f9tqQ9qXTthckRqH/4ii4A/5k/dXc+rVkYHioloUf6C9iPJ1uVlSz0vNjoE7BfPtitgdDepx8Ey1RHlIg==",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			var err error
			// TODO: Remove this. This is a really dumb hack to get tests working for now.
			if tc.method == http.MethodGet {
				time.Sleep(6 * time.Second)
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
			handler := http.HandlerFunc(HashHandler)
			handler.ServeHTTP(resRecorder, req)

			//Check status code
			status := resRecorder.Code
			if status != tc.statusCode {
				t.Errorf("Expected %v but instead got %v", tc.statusCode, status)
			}

			//Check body
			id := resRecorder.Body.String()
			if id != tc.expected {
				t.Errorf("Expected %v but instead got %v", tc.expected, id)
			}

		})
	}
}

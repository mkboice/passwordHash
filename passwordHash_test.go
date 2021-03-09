package main

import (
	"sync"
	"testing"
)

type SpySleeper struct {
	Calls int
}

func (s *SpySleeper) Sleep() {
	s.Calls++
}

func TestPasswordHashIncrementID(t *testing.T) {
	cases := []struct {
		name         string
		passwordHash *PasswordHash
		expected     int
	}{
		{
			name: "Increment ID from 0",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			expected: 1,
		},
		{
			name: "Increment ID from 41",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{},
				id:             41,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			expected: 42,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			tc.passwordHash.incrementID()

			// Check body
			if tc.passwordHash.id != tc.expected {
				t.Errorf("Expected %v but instead got %v", tc.expected, tc.passwordHash.id)
			}

		})
	}
}

func TestPasswordHashGetHashByID(t *testing.T) {
	cases := []struct {
		name         string
		id           int
		passwordHash *PasswordHash
		expected     string
		found        bool
	}{
		{
			name: "Get hash from invalid id",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			found:    false,
			expected: "",
		},
		{
			name: "Get Hash from valid id",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{1: "something", 2: "anotherhash"},
				id:             2,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			found:    true,
			expected: "anotherhash",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			hash, ok := tc.passwordHash.getHashByID(tc.passwordHash.id)
			if ok != tc.found {
				t.Errorf("Found: Expected %v but instead got %v", tc.found, ok)
			}

			if hash != tc.expected {
				t.Errorf("Hash: Expected %v but instead got %v", tc.expected, hash)
			}

		})
	}
}

func TestPasswordHashGetRequestStats(t *testing.T) {
	cases := []struct {
		name             string
		passwordHash     *PasswordHash
		expectedRequests int
		expectedTime     int
	}{
		{
			name: "Get requests stats with no data returns zeros",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			expectedRequests: 0,
			expectedTime:     0,
		},
		{
			name: "Get requests stats with valid data",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{},
				id:             33,
				totalTimeMicro: 7892,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			expectedRequests: 33,
			expectedTime:     239,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			numRequests, averageTimeMicroseconds := tc.passwordHash.getRequestStats()
			if numRequests != tc.expectedRequests {
				t.Errorf("Requests: Expected %v but instead got %v", tc.expectedRequests, numRequests)
			}

			if averageTimeMicroseconds != tc.expectedTime {
				t.Errorf("Time: Expected %v but instead got %v", tc.expectedTime, averageTimeMicroseconds)
			}

		})
	}
}

func TestPasswordHashComputeHash(t *testing.T) {
	cases := []struct {
		name         string
		passwordHash *PasswordHash
		password     string
		expectedHash string
	}{
		{
			name: "Compute hash with valid password",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{},
				id:             0,
				totalTimeMicro: 0,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			password:     "angryMonkey",
			expectedHash: "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==",
		},
		{
			name: "Comptue hash with other valid password",
			passwordHash: &PasswordHash{
				passwords:      map[int]string{},
				id:             33,
				totalTimeMicro: 7892,
				sleeper:        &SpySleeper{},
				waitGroup:      new(sync.WaitGroup),
			},
			password:     "something",
			expectedHash: "mD1D3f9tqQ9qXTthckRqH/4ii4A/5k/dXc+rVkYHioloUf6C9iPJ1uVlSz0vNjoE7BfPtitgdDepx8Ey1RHlIg==",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.passwordHash.computeHash(tc.password)
			tc.passwordHash.waitGroup.Wait()
			hash, found := tc.passwordHash.getHashByID(tc.passwordHash.id)
			if found != true {
				t.Errorf("Found: Expected %v but instead got %v", true, found)
			}

			if hash != tc.expectedHash {
				t.Errorf("Hash: Expected %v but instead got %v", tc.expectedHash, hash)
			}

		})
	}
}

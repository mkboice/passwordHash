package main

import "time"

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

package main

import (
	"github.com/go-co-op/gocron"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	ServerURLs = []*url.URL{
		ParseURL("http://127.0.0.1:5000/"),
		ParseURL("http://127.0.0.1:5001/"),
		ParseURL("http://127.0.0.1:5002/"),
		ParseURL("http://127.0.0.1:5003/"),
		ParseURL("http://127.0.0.1:5004/"),
	}
	HealthStatus = make([]bool, len(ServerURLs))
	Mu           sync.RWMutex
)

// Functions to perform health checks on servers

// Parses a string URL and returns a pointer to a url object which is required by rev proxy's Director func
func ParseURL(urlStr string) *url.URL {
	u, _ := url.Parse(urlStr)
	return u
}

// Go routine to run health checks on a separate thread
func RunHealthCheck() {
	// Utilizes cron jobs to run checkServerHealth every 2 seconds
	// Here we create a Cron Scheduler Object from the gocron library
	s := gocron.NewScheduler(time.Local)

	// We loop through every server URL and check each server's health
	for idx, server := range ServerURLs {
		// We want the scheduler to perform a given function, every 2 seconds
		// The Do method takes a function as an argument which is where we will put checkServerHealth
		s.Every(2).Second().Do(func() {
			// We run the function and then log the server's health
			CheckServerHealth(idx, server)
			if HealthStatus[idx] {
				log.Printf("'%s' is healthy!", server)
			} else {
				log.Printf("'%s' is not healthy!", server)
			}
		})
	}
	// cmd to start the func, use "<-" syntax if the function has a return value
	s.StartAsync()
}

// Pings each server to assess health status, updating shared state accordingly
func CheckServerHealth(index int, server *url.URL) {
	// We check server health by pinging the server and checking the response
	resp, err := http.Get(server.String())

	// Acquire a write lock to access HealthStatus atomically
	Mu.Lock()
	// Deferring delays the unlock until the function returns
	defer Mu.Unlock()

	// If we don't get a 200, it means the server is offline, set the health status to false
	if err != nil || resp.StatusCode != http.StatusOK {
		HealthStatus[index] = false
		// If we do get a 200, it means the server is online, set the health status to true
	} else {
		HealthStatus[index] = true
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func GetNextHealthyServer() *url.URL {
	// Acquire a read lock to access HealthStatus atomically
	Mu.RLock()
	// Deferring delays the unlock until the function returns
	defer Mu.RUnlock()

	// Loop to find the next healthy server
	for i := 0; i < len(ServerURLs); i++ {
		if HealthStatus[i] {
			return ServerURLs[i]
		}
	}
	// Fallback: if all servers are unhealthy, return a default server or error
	log.Printf("Error")
	return nil
}

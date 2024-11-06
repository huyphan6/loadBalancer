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

func ParseURL(urlStr string) *url.URL {
	u, _ := url.Parse(urlStr)
	return u
}

// Go routine to run health checks on a separate thread
func RunHealthCheck() {
	// Utilizes cron jobs to run checkServerHealth every 2 seconds
	// Here we create a Cron Scheduler Object called s
	s := gocron.NewScheduler(time.Local)

	// Looping syntax similar to Python
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
	// How to start the func, use "<-" if the function has a return value
	s.StartAsync()
}

func CheckServerHealth(index int, server *url.URL) {
	// We check server health by pinging the server and checking the response
	resp, err := http.Get(server.String())

	Mu.Lock()
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
	Mu.RLock()
	defer Mu.RUnlock()

	// Loop to find the next healthy server
	for i := 0; i < len(ServerURLs); i++ {
		if HealthStatus[i] {
			return ServerURLs[i]
		}
	}
	// Fallback: if all servers are unhealthy, return a default server or error
	return ServerURLs[0]
}
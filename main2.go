package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var sad = false
var busy = false

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	if name == "" {
		name = "guest"
	}
	// Set sad flag depending on current value and payload contents
	if sad {
		if strings.ToLower(name) == "happy" {
			sad = false
		}
	} else {
		if strings.ToLower(name) == "sad" {
			sad = true
		}
	}

	if strings.Contains(strings.ToLower(name), "busy") {
		// Flip if received "busy"
		busy = !busy
	}
	const MAX_LEN = 32
	guest_name, _ := os.LookupEnv("HOSTNAME")
	str_len := len(guest_name)
	from := 0
	if str_len > MAX_LEN {
		from = str_len - MAX_LEN
	}
	log.Printf("Received request for %s\n", name)
	w.Write([]byte(fmt.Sprintf("Hello, %s from container %s (sad: %t, busy: %t)\n", name, guest_name[from:str_len], sad, busy)))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	var status int
	var ret string
	if sad {
		status = http.StatusInternalServerError
		ret = fmt.Sprintf("error")
	} else {
		status = http.StatusOK
		ret = fmt.Sprintf("ok")
	}
	w.WriteHeader(status)
	w.Write([]byte(ret))
	log.Printf("HealthHandler sent %d\n", status)
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	var status int
	var ret string
	if busy {
		status = http.StatusInternalServerError
		ret = fmt.Sprintf("busy")
	} else {
		status = http.StatusOK
		ret = fmt.Sprintf("ok")
	}
	w.WriteHeader(status)
	w.Write([]byte(ret))
	log.Printf("ReadinessHandler sent %d\n", status)
}

func main() {
	// Create Server and Route Handlers
	r := mux.NewRouter()

	r.HandleFunc("/", handler)
	r.HandleFunc("/health", healthHandler)
	r.HandleFunc("/readiness", readinessHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}

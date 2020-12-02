package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestUserAgent(t *testing.T) {
	expectedUserAgent := "MyBrowser246"
	testOutput := "testing.png"
	userAgent := ""
	failed := false
	var srv http.Server

	// Start server
	go func() {
		srv = http.Server{
			Addr: "localhost:50000",
			Handler: http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				userAgent = req.Header["User-Agent"][0]
			}),
		}
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic("Server has failed")
		}
	}()

	// Call Program
	os.Args = []string{"cmd",
		"-uri=" + "http://localhost:50000/",
		"-user-agent=" + expectedUserAgent,
		"-output=" + testOutput,
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	main()

	// Assert
	log.Println("Server retreived User Agent:", userAgent)
	if userAgent != expectedUserAgent {
		failed = true
		log.Printf("User agent not set correctly, expected '%s', having '%s'", expectedUserAgent, userAgent)
	}

	// Cleanup
	os.Remove(testOutput)
	srv.Shutdown(context.Background())
	if failed {
		t.Fail()
	}
}

func TestTimeout(t *testing.T) {
	testOutput := "testing.png"
	failed := false

	// Start server
	go func() {
		if err := http.ListenAndServe("localhost:50000", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			time.Sleep(20 * time.Second)
		})); err != nil {
			panic("Server has failed")
		}
	}()

	// Call Program
	os.Args = []string{"cmd",
		"-uri=" + "http://localhost:50000/",
		"-timeout=10s",
		"-output=" + testOutput,
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	if !callMainRecovering() {
		log.Println("Timeout was not triggered")
		failed = true
	}

	// Cleanup
	os.Remove(testOutput)
	if failed {
		t.Fail()
	}
}

func callMainRecovering() (paniced bool) {
	defer func() {
		if p := recover(); p != nil {
			paniced = true
		}
	}()
	main()
	return
}

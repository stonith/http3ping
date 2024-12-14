package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"os"
	"io"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"crypto/tls"
)

func main() {
	// Define command-line flags
	url := flag.String("url", "", "The URL to send requests to")
	pause := flag.Int("pause", 1, "Pause duration between requests in seconds")
	count := flag.Int("count", 1, "Number of requests to send")
	inc := flag.Int("inc", 0, "Increment the pause length by this amount")
	keepalive := flag.Int("keepalive", 0, "Keepalive period in seconds")
	idleTimeout := flag.Int("idletimeout", 2000, "Idle timeout in seconds")
	flag.Parse()

	// Validate flags
	if *url == "" {
		log.Fatal("URL is required")
	}
	if *pause < 0 {
		log.Fatal("Pause duration must be non-negative")
	}
	if *count < 1 {
		log.Fatal("Count must be at least 1")
	}

	var keyLog io.Writer
	if filename := os.Getenv("SSLKEYLOGFILE"); len(filename) > 0 {
		f, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Could not create key log file: %s\n", err.Error())
			os.Exit(1)
		}
		defer f.Close()
		keyLog = f
	}


	quicConfig := &quic.Config{
		KeepAlivePeriod: time.Duration(*keepalive) * time.Second,
		MaxIdleTimeout:  time.Duration(*idleTimeout) * time.Second,
	}

	// Force QUIC for connections and don't verify TLS
	transport := &http3.Transport{
		QUICConfig: quicConfig,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			KeyLogWriter:       keyLog,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // Set request timeout to 10 seconds
	}

	// Send requests in a loop
	for i := 0; i < *count; i++ {
		now := time.Now().Format("2006-01-02 15:04:05.000") // time when request was originally initiated
		resp, err := client.Get(*url)
		if err != nil {
			log.Printf("Error making request: %v", err)
			continue
		}

		fmt.Printf("[%s] Request %d: %s\n", now, i+1, resp.Status)
		resp.Body.Close()

		// Pause between requests
		time.Sleep(time.Duration(*pause+i*(*inc)) * time.Second)
	}
}

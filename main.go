package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// APIVersion1 represents the REST API Version used in the URL path
const APIVersion1 = "/v1"
var c Config
var msgsStore MessageStore

func main() {
	ticker := time.NewTicker(time.Second)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Printf("Received signal %v, shutting down...", sig)
		done <- true
	}()

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				os.Exit(0)
			case <-ticker.C:
				log.Printf("Tick passed, sending data")

				if len(c.Peers) == 0 {
					log.Print("I have no peers, receiving only")
					break
				}

				for _, p := range c.Peers {
					// Prepare Message
					m := NewMessage(p)

					// Build URL
					peerURL, err := url.Parse(p + APIVersion1 + "/store/write")
					if err != nil {
						log.Printf("Could not parse peer URL: %s", err)
					}
					log.Printf("Contacting peer: %v", peerURL)

					// Marshal to json
					body, err := json.Marshal(m)
					if err != nil {
						log.Fatalf("Error: %v", err)
					}
					log.Printf("JSON body is: %s", body)

					req, err := http.NewRequest("POST", peerURL.String(), bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						log.Printf("Could not send to peer: %s", p)
					}
					log.Printf("Peer %v responded: %v", p, resp.Status)
				}
			}
		}
	}()

	err := envconfig.Process("GUZZLER", &c)
	if err != nil {
		log.Printf("Error parsing configuration from environment: %s", err)
	}
	log.Printf("Configuration readFromStore from environment: %+v", c)
	http.HandleFunc(APIVersion1+"/store/read", readFromStore)
	http.HandleFunc(APIVersion1+"/store/write", writeToStore)
	http.ListenAndServe(c.Bind, nil)
}

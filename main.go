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

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

// Starting some guzzlers:
// GUZZLER_NAME=guz_1 GUZZLER_BIND="localhost:8080" GUZZLER_PEERS="http://localhost:8081,http://localhost:8082" ./guzzler &
// GUZZLER_NAME=guz_1 GUZZLER_BIND="localhost:8081" GUZZLER_PEERS="http://localhost:8080,http://localhost:8082" ./guzzler &
// GUZZLER_NAME=guz_1 GUZZLER_BIND="localhost:8082" GUZZLER_PEERS="http://localhost:8080,http://localhost:8081" ./guzzler &

// Config manages the service configuration
type Config struct {
	Name  string   `envconfig:"GUZZLER_NAME" required:"true"`
	Bind  string   `envconfig:"GUZZLER_BIND" default:"127.0.0.1:8080" required:"false"`
	Peers []string `envconfig:"GUZZLER_PEERS" required:"true" split_words:"true"` /* ","-separated list of peers */
}

// Storing messages
type MessageStore struct {
	Messages []Message `json:messages`
}

// Message represents a payload for sending/receiving
type Message struct {
	ID      uuid.UUID `json:"id"`
	Created time.Time `json:"created"`
	Origin  string    `json:"origin"`
	Body    []byte    `json:"body"`
}

// APIVersion represents the REST API Version used in the URL path
const APIVersion = "/v1"

var c Config
var msgsStore MessageStore

// Store an incomming message
func store(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received: %v", r)

	decoder := json.NewDecoder(r.Body)
	var msg Message
	err := decoder.Decode(&msg)
	if err != nil {
		panic(err)
	}
	msgsStore.messages = append(msgsStore.messages, msg)
	log.Printf("Collected %d messages", len(msgsStore.messages))
	w.WriteHeader(http.StatusCreated)
}

// Return collected messages from other clients GET v1/read
func received(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(&msgsStore.messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

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
					m := Message{uuid.New(), time.Now(), c.Name, []byte("Blubb Body")}

					// Build URL
					peerURL, err := url.Parse(p + APIVersion + "/store")
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
	log.Printf("Configuration received from environment: %+v", c)
	http.HandleFunc(APIVersion+"/read", received)
	http.HandleFunc(APIVersion+"/store", store)
	http.ListenAndServe(c.Bind, nil)
}

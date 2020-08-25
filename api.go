package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Store an incomming message
func writeToStore(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received: %v", r)

	decoder := json.NewDecoder(r.Body)
	var msg Message
	err := decoder.Decode(&msg)
	if err != nil {
		panic(err)
	}
	msgsStore.Messages = append(msgsStore.Messages, msg)
	log.Printf("Collected %d messages", len(msgsStore.Messages))
	w.WriteHeader(http.StatusCreated)
}

// Return collected messages from other clients GET v1/read
func readFromStore(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(&msgsStore.Messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
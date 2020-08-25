package main

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

// Message represents a payload for sending/receiving
type Message struct {
	ID      uuid.UUID `json:"id"`
	Created time.Time `json:"created"`
	Origin  string    `json:"origin"`
	Body    string    `json:"body"`
}

func NewMessage(receiver string) *Message {
	qID := rand.Intn(len(Quotes))
	mb := Quotes[qID]
	m := Message{uuid.New(), time.Now(), receiver, mb}
	return &m
}
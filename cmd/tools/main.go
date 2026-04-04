package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Create JetStream Context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	// Define the Stream configuration
	streamName := "OUTBOX_EVENTS"
	cfg := &nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"event.*"}, // Matches your event.Type
		Storage:  nats.FileStorage,    // Critical for At-Least-Once

		// Retention and Discard Policy
		Retention: nats.LimitsPolicy, // Keep messages until limits reached
		Discard:   nats.DiscardOld,   // Drop oldest if space is needed

		// Deduplication: This works with your X-Event-ID header
		Duplicates: 2 * time.Minute,
	}

	// Add the stream
	_, err = js.AddStream(cfg)
	if err != nil {
		log.Fatalf("failed to create stream: %v", err)
	}

	log.Printf("Stream %s created successfully!", streamName)
}

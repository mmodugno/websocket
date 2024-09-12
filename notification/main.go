package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const oid = "333"

// Message structure
type (
	Message struct {
		Action  string      `json:"action"`
		Message MessageData `json:"message"`
		OrderID string      `json:"order_id"`
	}

	MessageData struct {
		ID      string `json:"id"`
		Number  int    `json:"number"`
		Message string `json:"message"`
		Date    string `json:"date"`
		OrderID string `json:"order_id"`
	}
)

func connectAndSendMessage() error {

	u := url.URL{Scheme: "wss", Host: "7w9u7ihj8e.execute-api.us-east-1.amazonaws.com", Path: "/production", RawQuery: "order_id=ABC"}
	log.Printf("Connecting to %s", u.String())

	// Establish a connection to the WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	// Close connection when finished
	defer conn.Close()

	// Create the message in the required format
	msg := Message{
		Action:  "notification",
		OrderID: oid,
		Message: MessageData{
			OrderID: oid,
			ID: func() string {
				u, _ := uuid.NewUUID()
				return u.String()
			}(),
			Number:  1,
			Message: "Holaaaaaaaaaaaaaa",
			Date:    time.Now().String(),
		},
	}

	// Convert the message to JSON
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send the JSON message
	err = conn.WriteMessage(websocket.TextMessage, msgJSON)
	if err != nil {
		return err
	}

	log.Printf("Sent message: %s", string(msgJSON))
	return nil
}

func main() {
	if err := connectAndSendMessage(); err != nil {
		log.Fatalf("Error: %v", err)
	} else {
		log.Println("Message sent successfully!")
	}
}

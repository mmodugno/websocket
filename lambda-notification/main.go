package main

import (
	"context"
	"encoding/json"
	"github.com/Bancar/lambda-go"
	"github.com/google/uuid"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type (
	Message struct {
		Action  string      `json:"action"`
		Message MessageData `json:"message"`
		OrderID string      `json:"order_id"`
	}

	MessageData struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Date    string `json:"date"`
		OrderID string `json:"order_id"`
	}
)

func main() {
	lambda.EnableLocalHTTP("8080")
	lambda.AsyncStart(Do, false)
}

func Do(_ context.Context, input *MessageData) error {
	u := url.URL{Scheme: "wss", Host: "o2hn4hxw55.execute-api.us-east-1.amazonaws.com", Path: "/dev", RawQuery: "order_id=default"}
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
		Action:  "sendmessage",
		OrderID: input.OrderID,
		Message: MessageData{
			OrderID: input.OrderID,
			ID: func() string {
				u, _ := uuid.NewUUID()
				return u.String()
			}(),
			Status: input.Status,
			Date:   time.Now().Format("2006-01-02 15:04:05"),
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

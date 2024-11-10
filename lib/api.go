package lib

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

type (
	RequestConnection struct {
		OrderID string `json:"order_id"`
	}
	Request struct {
		Action  string `json:"action"`
		OrderID string `json:"order_id"`
	}
	ResponseConnection struct {
		Message      string `json:"message"`
		ConnectionID string `json:"connectionId"`
		OrderID      string `json:"orderId"`
	}
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

func BuildResponse(status int, body interface{}) events.APIGatewayProxyResponse {
	responseBody, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf("invalid response")}
	}
	return events.APIGatewayProxyResponse{StatusCode: status, Body: string(responseBody)}
}

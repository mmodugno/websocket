package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"net/http"
)

func main() {
	lambda.Start(handler)
}

type (
	RequestBody struct {
		OrderID string `json:"order_id"`
		Action  string `json:"action"`
	}
	MessageData struct {
		ID      string `json:"id,omitempty"`
		Status  string `json:"status"`
		Date    string `json:"date,omitempty"`
		OrderID string `json:"order_id"`
	}
)

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Unable to load AWS config: %v", err)
		return createErrorResponse(500, fmt.Sprintf("AWS config error: %v", err)), nil
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)
	endpoint := fmt.Sprintf("https://%s/%s", request.RequestContext.DomainName, request.RequestContext.Stage)
	apigatewayclient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	var msg RequestBody
	err = json.Unmarshal([]byte(request.Body), &msg)
	if err != nil {
		log.Printf("Error parsing WebSocket message: %v", err)
		return createErrorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}
	if msg.OrderID == "" {
		log.Printf("empty order id")
		return createErrorResponse(http.StatusBadRequest, "Missing order_id"), nil
	}

	event, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"eventId": &types.AttributeValueMemberS{Value: msg.OrderID},
		},
		TableName: aws.String("WebSocketMessages"),
	})
	if err != nil {
		log.Printf("event not found: %v", err)
		return createErrorResponse(500, "cannot get item"), nil
	}

	if len(event.Item) == 0 { // Check if the item is empty
		log.Printf("Event not found for OrderID: %s", msg.OrderID)
		if err := sendMessage(ctx, apigatewayclient, request.RequestContext.ConnectionID,
			MessageData{
				Status:  "NOT FOUND",
				OrderID: msg.OrderID,
			}); err != nil {
			log.Printf("Failed to send message: %v", err)
			return createErrorResponse(500, "Failed to send WebSocket response"), nil
		}
		return createErrorResponse(404, "Event not found"), nil
	}

	date, ok := event.Item["date"].(*types.AttributeValueMemberS)
	if !ok {
		return createErrorResponse(500, "Missing date attribute"), nil
	}
	messageId, ok := event.Item["messageId"].(*types.AttributeValueMemberS)
	if !ok {
		return createErrorResponse(500, "Missing messageId attribute"), nil
	}
	status, ok := event.Item["status"].(*types.AttributeValueMemberS)
	if !ok {
		return createErrorResponse(500, "Missing status attribute"), nil
	}

	response := MessageData{
		ID:      messageId.Value,
		Status:  status.Value,
		Date:    date.Value,
		OrderID: msg.OrderID,
	}

	if err := sendMessage(ctx, apigatewayclient, request.RequestContext.ConnectionID, response); err != nil {
		log.Printf("Failed to send message: %v", err)
		return createErrorResponse(500, "Failed to send WebSocket response"), nil
	}

	responseBody, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func sendMessage(ctx context.Context, client *apigatewaymanagementapi.Client, connectionID string, message interface{}) error {
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	_, err = client.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &connectionID,
		Data:         messageData,
	})

	if err != nil {
		log.Printf("PostToConnection failed: %v", err)
		return fmt.Errorf("PostToConnection failed: %v", err)
	}

	return nil
}

func createErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       fmt.Sprintf(`{"error":"%s"}`, message),
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type (
	WebSocketMessage struct {
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

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Unable to load AWS config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf(`{"message":"%v"}`, err)}, nil
	}

	// Initialize DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)
	endpoint := fmt.Sprintf("https://%s/%s", event.RequestContext.DomainName, event.RequestContext.Stage)

	log.Printf("Complete URL: %s", endpoint)

	apigatewayclient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	var msg WebSocketMessage
	err = json.Unmarshal([]byte(event.Body), &msg)
	if err != nil {
		log.Printf("Error parsing WebSocket message: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	log.Printf("Received body: %s", msg)

	if msg.OrderID == "" {
		log.Printf("empty order id")
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}
	ttl := time.Now().Add(1 * time.Hour).Unix()
	// Prepare the item to store in DynamoDB
	item := map[string]types.AttributeValue{
		"eventId":   &types.AttributeValueMemberS{Value: msg.OrderID},
		"status":    &types.AttributeValueMemberS{Value: msg.Message.Status},
		"messageId": &types.AttributeValueMemberS{Value: msg.Message.ID},
		"date":      &types.AttributeValueMemberS{Value: msg.Message.Date},
		"ttl":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", ttl)},
	}

	// Insert into DynamoDB
	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("WebSocketMessages"),
		Item:      item,
	})
	if err != nil {
		log.Printf("DynamoDB error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"message":"Error saving message","error":"%v"}`, err),
		}, nil
	}

	// Query DynamoDB for connections with the specified order ID
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("WebSocketConnections"),
		IndexName:              aws.String("orderId-index"),
		KeyConditionExpression: aws.String("orderId = :orderID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":orderID": &types.AttributeValueMemberS{Value: msg.OrderID},
		},
	}

	connections, err := dynamoClient.Query(ctx, queryInput)
	if err != nil {
		log.Printf("DynamoDB query error: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `fmt.Sprintf({"message":"%v"}, err)`}, nil
	}

	log.Printf("Connections: %v", connections.Items)

	// Send the message to all other connections
	var sendMessages []error
	for _, item := range connections.Items {
		connectionID := item["connectionId"].(*types.AttributeValueMemberS).Value
		log.Printf("Connection id: %v", connectionID)
		// Avoid sending to the same connection that originated the message
		if connectionID != event.RequestContext.ConnectionID {
			err := sendMessage(apigatewayclient, connectionID, msg.Message)
			if err != nil {
				log.Printf("Failed to send message to connection %s: %v", connectionID, err)
				sendMessages = append(sendMessages, err)
			}
		}
	}
	if len(sendMessages) > 0 {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"message": "Message sent successfully"}`,
	}, nil
}

func sendMessage(client *apigatewaymanagementapi.Client, connectionID string, message interface{}) error {
	// Marshal the message into JSON format
	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Send the message using the PostToConnection API
	_, err = client.PostToConnection(context.TODO(), &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &connectionID,
		Data:         messageData,
	})

	if err != nil {
		log.Printf("PostToConnection failed: %v", err)
		return fmt.Errorf("PostToConnection failed: %v", err)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}

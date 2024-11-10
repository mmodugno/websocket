package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"net/http"
)

type (
	ACKMessage struct {
		Action  string `json:"action"`
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

	var msg ACKMessage
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

	_, err = dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"eventId": &types.AttributeValueMemberS{Value: msg.OrderID},
		},
		TableName: aws.String("WebSocketMessages"),
	})

	if err != nil {
		log.Printf("DynamoDB error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("cannot delete item"),
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"message": "Message sent successfully"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}

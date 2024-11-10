package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type RequestBody struct {
	OrderID string `json:"order_id"`
}

type Response struct {
	Message      string `json:"message"`
	ConnectionID string `json:"connectionId"`
	OrderID      string `json:"orderId"`
}

var dynamoClient *dynamodb.Client

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract orderId from query string or body
	orderID := request.QueryStringParameters["order_id"]
	if orderID == "" {
		var body RequestBody
		err := json.Unmarshal([]byte(request.Body), &body)
		if err != nil {
			log.Printf("Failed to parse request body: %v", err)
			return events.APIGatewayProxyResponse{
				StatusCode: 400,
				Body:       fmt.Sprintf(`{"message":"Invalid request body"}`),
			}, nil
		}
		orderID = body.OrderID
	}

	// Prepare the item to store in DynamoDB
	item := map[string]types.AttributeValue{
		"connectionId": &types.AttributeValueMemberS{Value: request.RequestContext.ConnectionID},
		"orderId":      &types.AttributeValueMemberS{Value: orderID},
	}

	// Insert into DynamoDB
	_, err := dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("WebSocketConnections"),
		Item:      item,
	})
	if err != nil {
		log.Printf("DynamoDB error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"message":"Error saving connection","error":"%v"}`, err),
		}, nil
	}

	// Return success response
	response := Response{
		Message:      "Connection saved",
		ConnectionID: request.RequestContext.ConnectionID,
		OrderID:      orderID,
	}
	responseBody, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}

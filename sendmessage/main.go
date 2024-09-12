package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws-lambda-go/lambda"
	"github.com/aws-sdk-go-v2/aws"
	"github.com/aws-sdk-go-v2/config"
	"github.com/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-lambda-go/events"
)

type RequestBody struct {
	Message string `json:"message"`
	OrderID string `json:"order_id"`
}

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Unable to load AWS config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf(`{"message":"%v"}`, err)}, nil
	}

	// Initialize DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Parse request body
	var body RequestBody
	err = json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		log.Printf("Error parsing request body: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: `{"message":"Invalid request body"}`}, nil
	}

	// Query DynamoDB for connections
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String("WebSocketConnections"),
		IndexName:              aws.String("orderId-index"),
		KeyConditionExpression: aws.String("orderId = :order_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":order_id": &types.AttributeValueMemberS{Value: body.OrderID},
		},
	}

	queryOutput, err := dynamoClient.Query(context.TODO(), queryInput)
	if err != nil {
		log.Printf("DynamoDB query error: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: fmt.Sprintf(`{"message":"%v"}`, err)}, nil
	}

	// Initialize API Gateway Management API client
	apiGatewayClient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.Endpoint = fmt.Sprintf("https://%s/%s", request.RequestContext.DomainName, request.RequestContext.Stage)
	})

	// Send messages to each connection
	var sendMessages []error
	for _, item := range queryOutput.Items {
		connectionID, ok := item["connectionId"].(*types.AttributeValueMemberS)
		if !ok || connectionID.Value == request.RequestContext.ConnectionID {
			continue
		}

		_, err = apiGatewayClient.PostToConnection(context.TODO(), &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionID.Value),
			Data:         []byte(body.Message),
		})
		if err != nil {
			sendMessages = append(sendMessages, err)
		}
	}

	// Check for errors in sending messages
	if len(sendMessages) > 0 {
		log.Printf("Errors sending messages: %v", sendMessages)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"message":"Failed to send some messages"}`}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: `{"message":"Message sent successfully"}`}, nil
}

func main() {
	lambda.Start(handler)
}

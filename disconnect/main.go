package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
)

func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize DynamoDB client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Prepare the key for deletion
	key := map[string]types.AttributeValue{
		"connectionId": &types.AttributeValueMemberS{Value: request.RequestContext.ConnectionID},
	}

	// Delete the item from DynamoDB
	_, err = dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("WebSocketConnections"),
		Key:       key,
	})
	if err != nil {
		log.Printf("DynamoDB error: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"message":"Error deleting connection","error":"%v"}`, err),
		}, nil
	}

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"message":"Connection deleted"}`,
	}, nil
}

func main() {
	lambda.Start(handler)
}

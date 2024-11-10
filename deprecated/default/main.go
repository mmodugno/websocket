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
	"log"
)

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	connectionID := event.RequestContext.ConnectionID

	// Initialize API Gateway Management API client
	endpoint := fmt.Sprintf("https://%s/%s", event.RequestContext.DomainName, event.RequestContext.Stage)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	client := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.EndpointResolver = apigatewaymanagementapi.EndpointResolverFromURL(endpoint)
	})

	// Get connection info
	getConnInput := &apigatewaymanagementapi.GetConnectionInput{
		ConnectionId: aws.String(connectionID),
	}

	connectionInfo, err := client.GetConnection(ctx, getConnInput)
	if err != nil {
		log.Printf("Error getting connection: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Attach connectionID to the response data
	connectionData := map[string]interface{}{
		"connectionID":   connectionID,
		"connectionInfo": connectionInfo,
	}
	jsonData, err := json.Marshal(connectionData)
	if err != nil {
		log.Printf("Error marshaling connection data: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// Send message back to the WebSocket client
	postToConnInput := &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         []byte("Use the sendmessage route to send a message. Your info: " + string(jsonData)),
	}

	_, err = client.PostToConnection(ctx, postToConnInput)
	if err != nil {
		log.Printf("Error posting to connection: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}

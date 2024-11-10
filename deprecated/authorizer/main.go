package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Handler for the WebSocket Authorizer
func Handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (map[string]interface{}, error) {
	// Extract authorization token from query parameters
	authToken := request.QueryStringParameters["Authorization"]

	// Simple token validation logic
	if authToken == "valid-token" {
		return generatePolicy("user123", "Allow", request.RequestContext.Stage)
	}

	return generatePolicy("user123", "Deny", request.RequestContext.Stage)
}

// Helper function to generate IAM Policy
func generatePolicy(principalID, effect, stage string) (map[string]interface{}, error) {
	// Construct the resource ARN for the WebSocket connection
	resourceArn := fmt.Sprintf("arn:aws:execute-api:*:*:*/%s/*", stage)

	// Create the IAM policy statement
	statement := map[string]interface{}{
		"Action":   "execute-api:Invoke",
		"Effect":   effect,
		"Resource": []string{resourceArn},
	}

	// Create the policy document
	policyDocument := map[string]interface{}{
		"Version":   "2012-10-17",
		"Statement": []interface{}{statement},
	}

	// Create the final response
	response := map[string]interface{}{
		"principalId":    principalID,
		"policyDocument": policyDocument,
		"context": map[string]interface{}{
			"user": principalID, // Custom context information
		},
	}

	return response, nil
}

func main() {
	lambda.Start(Handler)
}

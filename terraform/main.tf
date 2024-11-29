provider "aws" {
  region = "us-east-1"  # Specify your AWS region
  access_key = "-"
  secret_key = "-"
  token      = "-"
}

# Data block to reference the existing Lambda functions
data "aws_lambda_function" "existing_connect_lambda" {
  function_name = "WebsocketConnectTest"  # Replace with the name of your existing connect Lambda function
}

data "aws_lambda_function" "existing_disconnect_lambda" {
  function_name = "WebsocketDisconnectTest"  # Replace with the name of your existing disconnect Lambda function
}

data "aws_lambda_function" "existing_request_lambda" {
  function_name = "WebsocketRequestTest"  # Replace with the name of your existing disconnect Lambda function
}

data "aws_lambda_function" "existing_ack_lambda" {
  function_name = "WebsocketAckTest"  # Replace with the name of your existing disconnect Lambda function
}

# API Gateway WebSocket API
resource "aws_apigatewayv2_api" "websocket_api" {
  name                       = "websocket-api-test-terra"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}

# Connect Route for WebSocket
resource "aws_apigatewayv2_route" "connect_route" {
  api_id    = aws_apigatewayv2_api.websocket_api.id
  route_key = "$connect"

  target = "integrations/${aws_apigatewayv2_integration.connect_integration.id}"
}

# Disconnect Route for WebSocket
resource "aws_apigatewayv2_route" "disconnect_route" {
  api_id    = aws_apigatewayv2_api.websocket_api.id
  route_key = "$disconnect"

  target = "integrations/${aws_apigatewayv2_integration.disconnect_integration.id}"
}

# Request Route for WebSocket
resource "aws_apigatewayv2_route" "request_route" {
  api_id    = aws_apigatewayv2_api.websocket_api.id
  route_key = "request"

  target = "integrations/${aws_apigatewayv2_integration.request_integration.id}"
}

# Request Route for WebSocket
resource "aws_apigatewayv2_route" "ack_route" {
  api_id    = aws_apigatewayv2_api.websocket_api.id
  route_key = "ack"
  target = "integrations/${aws_apigatewayv2_integration.request_integration.id}"
}


# WebSocket API Gateway integration with existing Lambda for connect
resource "aws_apigatewayv2_integration" "connect_integration" {
  api_id          = aws_apigatewayv2_api.websocket_api.id
  integration_uri = data.aws_lambda_function.existing_connect_lambda.invoke_arn
  integration_type = "AWS_PROXY"
  integration_method = "POST"
}

# WebSocket API Gateway integration with existing Lambda for disconnect
resource "aws_apigatewayv2_integration" "disconnect_integration" {
  api_id          = aws_apigatewayv2_api.websocket_api.id
  integration_uri = data.aws_lambda_function.existing_disconnect_lambda.invoke_arn
  integration_type = "AWS_PROXY"
  integration_method = "POST"
}

# WebSocket API Gateway integration with existing Lambda for request
resource "aws_apigatewayv2_integration" "request_integration" {
  api_id          = aws_apigatewayv2_api.websocket_api.id
  integration_uri = data.aws_lambda_function.existing_request_lambda.invoke_arn
  integration_type = "AWS_PROXY"
  integration_method = "POST"
}

# WebSocket API Gateway integration with existing Lambda for request
resource "aws_apigatewayv2_integration" "ack_integration" {
  api_id          = aws_apigatewayv2_api.websocket_api.id
  integration_uri = data.aws_lambda_function.existing_ack_lambda.invoke_arn
  integration_type = "AWS_PROXY"
  integration_method = "POST"
}

# API Gateway Deployment
resource "aws_apigatewayv2_deployment" "websocket_deployment" {
  api_id = aws_apigatewayv2_api.websocket_api.id

  depends_on = [
    aws_apigatewayv2_route.connect_route,
    aws_apigatewayv2_route.disconnect_route,
    aws_apigatewayv2_route.request_route
  ]
}

# WebSocket API Stage
resource "aws_apigatewayv2_stage" "websocket_stage" {
  api_id      = aws_apigatewayv2_api.websocket_api.id
  name        = "dev"
  auto_deploy = true # Enable this if you want changes to auto-deploy
}

# Lambda Permission to allow API Gateway to invoke the existing connect function
resource "aws_lambda_permission" "apigw_connect_lambda_permission" {
  statement_id  = "AllowExecutionFromAPIGatewayConnect"
  action        = "lambda:InvokeFunction"
  function_name = data.aws_lambda_function.existing_connect_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.websocket_api.execution_arn}/*/*"
}

# Lambda Permission to allow API Gateway to invoke the existing disconnect function
resource "aws_lambda_permission" "apigw_disconnect_lambda_permission" {
  statement_id  = "AllowExecutionFromAPIGatewayDisconnect"
  action        = "lambda:InvokeFunction"
  function_name = data.aws_lambda_function.existing_disconnect_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.websocket_api.execution_arn}/*/*"
}

# Lambda Permission to allow API Gateway to invoke the existing request function
resource "aws_lambda_permission" "apigw_request_lambda_permission" {
  statement_id  = "AllowExecutionFromAPIGatewayRequest"
  action        = "lambda:InvokeFunction"
  function_name = data.aws_lambda_function.existing_request_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.websocket_api.execution_arn}/*/*"
}

# Lambda Permission to allow API Gateway to invoke the existing request function
resource "aws_lambda_permission" "apigw_ack_lambda_permission" {
  statement_id  = "AllowExecutionFromAPIGatewayRequest"
  action        = "lambda:InvokeFunction"
  function_name = data.aws_lambda_function.existing_ack_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.websocket_api.execution_arn}/*/*"
}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}

data "aws_caller_identity" "current" {}

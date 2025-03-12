package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaHandler(ctx context.Context, event json.RawMessage) {
	log.Print("Inside lambdaHandler func")
	log.Printf("Context: %v\n", ctx)
	log.Printf("Event: %v\n", string(event))
}

func main() {
	log.Print("Inside main func")
	lambda.Start(lambdaHandler)
}

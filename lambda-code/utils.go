package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type Response struct {
	Status     bool   `json:"status"`
	Data       any    `json:"data"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func response(statusCode int, status bool, message string, data any) events.APIGatewayProxyResponse {
	log.Print("Inside response func")

	res := Response{
		Status:     status,
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
	log.Printf("res: %v", res)
	jsonRes, err := json.Marshal(res)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}
	}

	log.Printf("jsonRes: %v", jsonRes)
	log.Printf("string(jsonRes): %v", string(jsonRes))

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(jsonRes),
	}
}

func getHeaders(headers map[string]string, key string) string {
	for k, v := range headers {
		if strings.ToLower(k) == strings.ToLower(key) {
			return v
		}
	}
	return ""
}

func generateUUID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		log.Printf("Error generating uuid: %v", err)
		return "", err
	}

	return id.String(), err
}

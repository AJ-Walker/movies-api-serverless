package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	Status     bool   `json:"status"`
	Data       any    `json:"data"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

const (
	AWS_REGION string = "ap-south-1"
	TABLE_NAME string = "Movies"
	MODEL_ID   string = "anthropic.claude-3-sonnet-20240229-v1:0"
)

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside lambdaHandler func")
	log.Printf("Context: %v\n", ctx)
	log.Printf("Event: %v\n", event)

	log.Printf("Resource: %v\n", event.Resource)
	log.Printf("Path: %v\n", event.Path)
	log.Printf("Query Params: %v\n", event.QueryStringParameters)

	switch event.Path {
	case "/api/movies":
		// call movies func

		if year, ok := event.QueryStringParameters["year"]; ok {
			return getMoviesByYear(year)
		} else {
			return getMovies()
		}

	case "/api/movies/summary":
		// call movies summary func

		movieId, ok := event.QueryStringParameters["movieId"]

		if !ok {
			return response(http.StatusNotFound, false, "movieId query param missing", nil), nil
		}
		return getMovieSummary(movieId)
	}
	return response(http.StatusInternalServerError, false, "Wrong path provided", nil), nil
}

func init() {
	Init_DB()
	Init_Bedrock()
}

func main() {
	log.Print("Inside main func")
	lambda.Start(HandleRequest)
}

func getMovies() (events.APIGatewayProxyResponse, error) {
	log.Print("Inside getMovies func")

	result, err := GetAllMovies_DB()
	if err != nil {
		log.Print(err)
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	if len(result) == 0 {
		return response(http.StatusOK, false, "No movies found", nil), nil
	}

	return response(http.StatusOK, true, "Movies fetched successfully.", result), nil
}

func getMoviesByYear(year string) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside getMoviesByYear func")
	if year == "" {
		return response(http.StatusBadRequest, false, "year field missing", nil), nil
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Print(err)
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	result, err := GetMoviesByYear_DB(int16(yearInt))
	if err != nil {
		log.Print(err)
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	if len(result) == 0 {
		return response(http.StatusOK, false, "No movies found", nil), nil
	}

	return response(http.StatusOK, true, "Movies fetched successfully.", result), nil
}

func getMovieSummary(movieId string) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside getMoviesSummary func")

	if movieId == "" {
		return response(http.StatusBadRequest, false, "movieId cannot be empty", nil), nil
	}

	result, err := GetMovieSummary_DB(movieId)
	if err != nil {
		log.Print(err)
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	data := map[string]string{
		"summary": result,
	}
	return response(http.StatusOK, true, "Movie summary fetched.", data), nil
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

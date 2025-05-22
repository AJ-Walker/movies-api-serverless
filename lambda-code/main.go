package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	AWS_REGION  string = "ap-south-1"
	TABLE_NAME  string = "Movies"
	MODEL_ID    string = "anthropic.claude-3-sonnet-20240229-v1:0"
	BUCKET_NAME string = "movies-api-data"
)

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside lambdaHandler func")
	log.Printf("Context: %v\n", ctx)
	// log.Printf("Event: %v\n", event)

	// request, err := json.Marshal(event)
	// if err != nil {
	// 	log.Printf("%s", err.Error())
	// }

	// // log.Printf("Request: %v", string(request))

	log.Printf("Resource: %v\n", event.Resource)
	log.Printf("Path: %v\n", event.Path)
	log.Printf("Query Params: %v\n", event.QueryStringParameters)

	log.Printf("Body: %v\n", event.Body)
	log.Printf("Headers: %v\n", event.Headers)
	log.Printf("IsBase64Encoded: %v\n", event.IsBase64Encoded)

	switch {
	case event.Path == "/api/movies" && event.HTTPMethod == "GET":
		// movies related apis

		if year, ok := event.QueryStringParameters["year"]; ok {
			return getMoviesByYear(year)
		} else if movieId, ok := event.QueryStringParameters["movieId"]; ok {
			return getMovieById(movieId)
		} else {
			return getMovies()
		}

	case event.Path == "/api/movies" && event.HTTPMethod == "POST":
		// Add movie api

		contentType := getHeaders(event.Headers, "Content-Type")
		if contentType == "" {
			return response(http.StatusBadRequest, false, "Missing Content-Type header", nil), nil
		}

		mediaType, params, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "multipart/form-data" {
			log.Printf("Invalid Content-Type or parsing failed: %v", err)
			return response(http.StatusBadRequest, false, "Invalid or unsupported Content-Type", nil), nil
		}

		boundary := params["boundary"]
		if boundary == "" {
			log.Print("Boundary not found in Content-Type")
			return response(http.StatusBadRequest, false, "Missing boundary in Content-Type header", nil), nil
		}
		log.Printf("boundary: %v", boundary)

		bodyBytes, err := base64.StdEncoding.DecodeString(event.Body)
		if err != nil {
			log.Printf("failed to decode body: %v", err)
			return response(http.StatusBadRequest, false, fmt.Sprintf("failed to decode body: %v", err), nil), nil
		}

		bytesReader := bytes.NewReader(bodyBytes)
		multipartReader := multipart.NewReader(bytesReader, boundary)
		form, err := multipartReader.ReadForm(10 << 20) // Max 10MB
		if err != nil {
			log.Printf("Error parsing multipart form: %v", err)
			return response(http.StatusBadRequest, false, fmt.Sprintf("Error parsing form data: %v", err), nil), nil
		}

		log.Printf("Form Fields: %v", form.Value)
		log.Printf("Form Files: %v", form.File)

		return addMovie(form)

	case event.Path == "/api/movies" && event.HTTPMethod == "PUT":
		// Update existing movie api

		if movieId, ok := event.QueryStringParameters["movieId"]; ok {
			contentType := getHeaders(event.Headers, "Content-Type")
			if contentType == "" {
				return response(http.StatusBadRequest, false, "Missing Content-Type header", nil), nil
			}

			mediaType, params, err := mime.ParseMediaType(contentType)
			if err != nil || mediaType != "multipart/form-data" {
				log.Printf("Invalid Content-Type or parsing failed: %v", err)
				return response(http.StatusBadRequest, false, "Invalid or unsupported Content-Type", nil), nil
			}

			boundary := params["boundary"]
			if boundary == "" {
				log.Print("Boundary not found in Content-Type")
				return response(http.StatusBadRequest, false, "Missing boundary in Content-Type header", nil), nil
			}
			log.Printf("boundary: %v", boundary)

			bodyBytes, err := base64.StdEncoding.DecodeString(event.Body)
			if err != nil {
				log.Printf("failed to decode body: %v", err)
				return response(http.StatusBadRequest, false, fmt.Sprintf("failed to decode body: %v", err), nil), nil
			}

			bytesReader := bytes.NewReader(bodyBytes)
			multipartReader := multipart.NewReader(bytesReader, boundary)
			form, err := multipartReader.ReadForm(10 << 20) // Max 10MB
			if err != nil {
				log.Printf("Error parsing multipart form: %v", err)
				return response(http.StatusBadRequest, false, fmt.Sprintf("Error parsing form data: %v", err), nil), nil
			}

			log.Printf("Form Fields: %v", form.Value)
			log.Printf("Form Files: %v", form.File)

			return updateMovie(movieId, form)
		} else {
			return response(http.StatusNotFound, false, "movieId query param missing", nil), nil
		}

	case event.Path == "/api/movies" && event.HTTPMethod == "DELETE":
		// Delete movie by Id

		if movieId, ok := event.QueryStringParameters["movieId"]; ok {
			return deleteMovie(movieId)
		} else {
			return response(http.StatusNotFound, false, "movieId query param missing", nil), nil
		}

	case event.Path == "/api/movies/summary" && event.HTTPMethod == "GET":
		// movies summary related apis

		if movieId, ok := event.QueryStringParameters["movieId"]; ok {
			return getMovieSummary(movieId)
		} else {
			return response(http.StatusNotFound, false, "movieId query param missing", nil), nil
		}
	}
	return response(http.StatusInternalServerError, false, "Wrong path provided", nil), nil
}

func init() {
	Init_DB()
	Init_Bedrock()
	Init_S3()
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
		return response(http.StatusNotFound, false, "No movies found", nil), nil
	}

	return response(http.StatusOK, true, "Movies fetched successfully.", result), nil
}

func getMoviesByYear(year string) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside getMoviesByYear func")
	if year == "" {
		return response(http.StatusBadRequest, false, "year cannot be empty", nil), nil
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

func getMovieById(movieId string) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside getMovieById func")

	if movieId == "" {
		return response(http.StatusBadRequest, false, "movieId cannot be empty", nil), nil
	}

	movie, err := GetMovieById_DB(movieId)

	if err != nil {
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}
	return response(http.StatusOK, true, "Movie fetched successfully", movie), nil
}

func addMovie(form *multipart.Form) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside addMovie func")

	if len(form.Value["title"]) == 0 || len(form.Value["releaseYear"]) == 0 || len(form.Value["genre"]) == 0 {
		return response(http.StatusBadRequest, false, "'title' and 'releaseYear' and 'genre' fields are required", nil), nil
	}

	title := form.Value["title"][0]
	releaseYear := form.Value["releaseYear"][0]
	genre := form.Value["genre"][0]

	if title == "" || releaseYear == "" || genre == "" {
		return response(http.StatusBadRequest, false, "'title' or 'releaseYear' or 'genre' field cannot be empty", nil), nil
	}

	// check if movie is being created with same title
	result, _ := GetMovieByTitle_DB(title)

	if strings.Trim(strings.ToLower(result.Title), " ") == strings.Trim(strings.ToLower(title), " ") {
		log.Printf("Movie with same title already exists")
		return response(http.StatusBadRequest, false, "movie with same title already exists", nil), nil
	}

	var objectUrl string

	movieId, err := generateUUID()
	if err != nil {
		return response(http.StatusBadRequest, false, "Error generating unique id", nil), nil
	}

	// check if movie image is provided
	if len(form.File) != 0 && len(form.File["coverImage"]) != 0 {
		coverImage := form.File["coverImage"][0]

		log.Printf("Movie coverImage file provided, Filename: %v", coverImage.Filename)

		fileExtension := filepath.Ext(coverImage.Filename)

		// upload file to s3
		key := fmt.Sprintf("%v%v", movieId, fileExtension)
		log.Printf("object key: %v", key)
		// key := fmt.Sprintf("%v-%v%v", strings.ReplaceAll(strings.ToLower(title), " ", "-"), releaseYear, fileExtension)
		// log.Printf("object key: %v", key)

		var err error
		objectUrl, err = PutObject_S3(coverImage, key)

		if err != nil {
			return response(http.StatusBadRequest, false, err.Error(), nil), nil
		}

		log.Printf("Object Url: %v", objectUrl)
	}

	year, err := strconv.Atoi(releaseYear)
	if err != nil {
		return response(http.StatusBadRequest, false, "Error converting string to int", nil), nil
	}

	movie := Movie{
		MovieId:     movieId,
		Title:       title,
		ReleaseYear: uint16(year),
		Genre:       genre,
	}

	if objectUrl != "" {
		movie.CoverUrl = objectUrl
	}

	if err := AddMovie_DB(movie); err != nil {
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	return response(http.StatusOK, true, "Movie added successfully", nil), nil
}

func updateMovie(movieId string, form *multipart.Form) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside updateMovie func")
	if movieId == "" {
		return response(http.StatusBadRequest, false, "movieId cannot be empty", nil), nil
	}

	if len(form.Value["title"]) == 0 || len(form.Value["releaseYear"]) == 0 || len(form.Value["genre"]) == 0 {
		return response(http.StatusBadRequest, false, "'title' and 'releaseYear' and 'genre' fields are required", nil), nil
	}

	title := form.Value["title"][0]
	releaseYear := form.Value["releaseYear"][0]
	genre := form.Value["genre"][0]

	if title == "" || releaseYear == "" || genre == "" {
		return response(http.StatusBadRequest, false, "'title' or 'releaseYear' or 'genre' field cannot be empty", nil), nil
	}

	// Check if movie exists with the provided movieId
	movie, err := GetMovieById_DB(movieId)
	if err != nil {
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	log.Print(movie)

	if strings.Trim(strings.ToLower(movie.Title), " ") != strings.Trim(strings.ToLower(title), " ") {
		// check if movie is being updated with same title
		result, _ := GetMovieByTitle_DB(title)

		if strings.Trim(strings.ToLower(result.Title), " ") == strings.Trim(strings.ToLower(title), " ") {
			log.Printf("Movie with same title already exists")
			return response(http.StatusBadRequest, false, "movie with same title already exists", nil), nil
		}
	}

	var objectUrl string

	// check if movie image is provided and update the existing with new
	if len(form.File) != 0 && len(form.File["coverImage"]) != 0 {
		coverImage := form.File["coverImage"][0]

		log.Printf("Movie coverImage file provided, Filename: %v", coverImage.Filename)

		fileExtension := filepath.Ext(coverImage.Filename)

		// upload file to s3
		key := fmt.Sprintf("%v%v", movie.MovieId, fileExtension)
		log.Printf("object key: %v", key)
		// key := fmt.Sprintf("%v-%v%v", strings.ReplaceAll(strings.ToLower(title), " ", "-"), releaseYear, fileExtension)
		// log.Printf("object key: %v", key)

		var err error
		objectUrl, err = PutObject_S3(coverImage, key)

		if err != nil {
			return response(http.StatusBadRequest, false, err.Error(), nil), nil
		}

		log.Printf("Object Url: %v", objectUrl)
	}

	// Convert releaseYear string into int
	year, err := strconv.Atoi(releaseYear)
	if err != nil {
		log.Print("Error converting releaseYear string into int")
		return response(http.StatusBadRequest, false, "Error converting releaseYear string into int", nil), nil
	}

	movie = Movie{
		Title:       title,
		ReleaseYear: uint16(year),
		Genre:       genre,
	}

	if objectUrl != "" {
		movie.CoverUrl = objectUrl
	}

	if err := UpdateMovieById_DB(movieId, movie); err != nil {
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	return response(http.StatusOK, true, "Movie updated successfully", nil), nil
}

func deleteMovie(movieId string) (events.APIGatewayProxyResponse, error) {
	log.Print("Inside deleteMovie func")
	if movieId == "" {
		return response(http.StatusBadRequest, false, "movieId cannot be empty", nil), nil
	}

	movie, err := DeleteMovieById_DB(movieId)
	if err != nil {
		return response(http.StatusBadRequest, false, err.Error(), nil), nil
	}

	if movie.CoverUrl != "" {

		splittedString := strings.Split(movie.CoverUrl, "/")

		objectKey := splittedString[len(splittedString)-1]
		log.Printf("ObjectKey: %v", objectKey)
		if err := DeleteObject_S3(objectKey); err != nil {
			log.Printf("Error while deleting object: %v", err)
		}
	}

	return response(http.StatusOK, true, "Movie deleted successfully", nil), nil
}

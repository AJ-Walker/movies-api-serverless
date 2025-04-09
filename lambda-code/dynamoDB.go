package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Movie struct {
	MovieId          string `json:"movieId" dynamodbav:"movieId"`
	Title            string `json:"title" dynamodbav:"title"`
	ReleaseYear      uint16 `json:"releaseYear" dynamodbav:"releaseYear"`
	Genre            string `json:"genre" dynamodbav:"genre"`
	CoverUrl         string `json:"coverUrl" dynamodbav:"coverUrl"`
	GeneratedSummary string `json:"generatedSummary,omitempty" dynamodbav:"generatedSummary,omitempty"`
}

var DynamoClient *dynamodb.Client

func Init_DB() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AWS_REGION))
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}
	DynamoClient = dynamodb.NewFromConfig(cfg)
}

func GetAllMovies_DB() ([]Movie, error) {
	log.Print("Inside GetAllMovies_DB func")

	result, err := DynamoClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(TABLE_NAME),
	})
	if err != nil {
		return nil, err
	}

	var movies []Movie
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

func GetMoviesByYear_DB(year int16) ([]Movie, error) {
	log.Print("Inside GetMoviesByYear_DB func")

	keyEx := expression.Key("releaseYear").Equal(expression.Value(year))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()

	if err != nil {
		return nil, err
	}

	result, err := DynamoClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:                 aws.String(TABLE_NAME),
		FilterExpression:          expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	log.Print(result)
	if err != nil {
		return nil, err
	}
	var movies []Movie
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &movies); err != nil {
		return nil, err
	}
	log.Print(movies)
	return movies, nil
}

func GetMovieSummary_DB(movieId string) (string, error) {
	log.Print("Inside GetMovieSummary_DB func")

	result, err := DynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"movieId": &types.AttributeValueMemberS{
				Value: movieId,
			},
		},
	})

	if err != nil {
		log.Print(err)
		return "", err
	}

	if len(result.Item) == 0 {
		log.Println("No movie found")
		return "", fmt.Errorf("No movie found")
	}
	var movie Movie
	if err := attributevalue.UnmarshalMap(result.Item, &movie); err != nil {
		log.Printf("Couldn't unmarshall update response. Here's why: %v\n", err)
		return "", err
	}

	if movie.GeneratedSummary == "" {
		log.Print("No summary available. Generate a summary.")

		movieSummary, err := GenerateMovieSummary(movie)
		if err != nil {
			log.Print(err)
			return "", err
		}

		// Save the summary for next time fetch for the movie
		if err := UpdateMovieSummary_DB(movie.MovieId, movieSummary); err != nil {
			log.Print(err)
			return "", err
		}

		return movieSummary, nil

	}
	return movie.GeneratedSummary, nil
}

func UpdateMovieSummary_DB(movieId string, summary string) error {
	log.Print("Inside UpdateMovieSummary_DB func")

	updateExpr := expression.Set(expression.Name("generatedSummary"), expression.Value(summary))
	expr, err := expression.NewBuilder().WithUpdate(updateExpr).Build()

	if err != nil {
		log.Printf("Couldn't build expression for update. Here's why: %v\n", err)
		return err
	} else {
		result, err := DynamoClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
			TableName: aws.String(TABLE_NAME),
			Key: map[string]types.AttributeValue{
				"movieId": &types.AttributeValueMemberS{Value: movieId},
			},
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ReturnValues:              types.ReturnValueAllNew,
		})

		if err != nil {
			log.Print(err)
			return err
		}
		var movie Movie
		if err := attributevalue.UnmarshalMap(result.Attributes, &movie); err != nil {
			log.Printf("Couldn't unmarshall update response. Here's why: %v\n", err)
			return err
		}
		log.Printf("Updated Movie: %v", movie)
		return nil
	}
}

func GetMovieById_DB(movieId string) (Movie, error) {
	log.Print("Inside GetMovieById_DB func")

	result, err := DynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"movieId": &types.AttributeValueMemberS{
				Value: movieId,
			},
		},
	})

	if err != nil {
		log.Printf("failed to get item from DynamoDB: %v", err)
		return Movie{}, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	if len(result.Item) == 0 {
		log.Print("No movie found")
		return Movie{}, fmt.Errorf("No movie found")
	}

	var movie Movie
	if err := attributevalue.UnmarshalMap(result.Item, &movie); err != nil {
		log.Printf("Couldn't unmarshall update response. Here's why: %v\n", err)
		return Movie{}, err
	}

	return movie, nil
}

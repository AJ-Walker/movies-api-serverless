package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	bedrockTypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ClaudeRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type ClaudeResponse struct {
	Completion string `json:"completion"`
}

func ListObjects_S3() {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	s3_client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Region = AWS_REGION
	})
	output, err := s3_client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(BUCKET_NAME),
	})
	if err != nil {
		fmt.Println(err)
	}

	for _, obj := range output.Contents {
		url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", BUCKET_NAME, AWS_REGION, aws.ToString(obj.Key))
		fmt.Printf("url=%s\n", url)
	}
}

func PutItems_DynamoDB(movies []Movie) error {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	var writeReqs []types.WriteRequest

	for _, movie := range movies {
		item, err := attributevalue.MarshalMap(movie)
		if err != nil {
			fmt.Println(err)
			return err
		}

		writeReqs = append(writeReqs, types.WriteRequest{PutRequest: &types.PutRequest{Item: item}})
	}

	// dynamoDb_Client.Scan(context.TODO(), &dynamodb.ScanInput{
	// 	TableName: aws.String("aws"),
	// })

	batchOutput, err := dynamoDb_Client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			TABLE_NAME: writeReqs,
		},
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(batchOutput)
	return nil
}

func GetMovies() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	result, err := dynamoDb_Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(TABLE_NAME),
	})

	var movies []Movie
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &movies); err != nil {
		fmt.Println(err)
	}

	fmt.Println(movies)

	return nil

}

func GetMoviesByYear(year int16) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	keyEx := expression.Key("releaseYear").Equal(expression.Value(year))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()

	if err != nil {
		fmt.Println(err)
	}

	result, err := dynamoDb_Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:                 aws.String(TABLE_NAME),
		FilterExpression:          expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	if err != nil {
		fmt.Println(err)
	}
	var movies []Movie
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &movies); err != nil {
		fmt.Println(err)
	}

	fmt.Println(movies)
	return nil
}

func GetMovieSummary(movieId string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	keyEx := expression.Key("movieId").Equal(expression.Value(movieId))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()

	result, err := dynamoDb_Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(TABLE_NAME),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	},
	)

	if err != nil {
		fmt.Println(err)
	}
	if len(result.Items) == 0 {
		fmt.Println("empty list")
	}
	var movie Movie
	if err := attributevalue.UnmarshalMap(result.Items[0], &movie); err != nil {
		fmt.Println(err)
	}

	if movie.GeneratedSummary == "" {
		fmt.Print("no summary available\n")

		// summary, err := generateSummary("")

		// if err != nil {
		// 	fmt.Println(err)
		// }

		// fmt.Printf("summary: %v", summary)

		// if summary != "" {

		// }
	}

	fmt.Println(movie)
	return nil
}

func generateSummary(prompt string) error {
	fmt.Println("Inside generateSummary func")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	bedrockRuntimeClient := bedrockruntime.NewFromConfig(cfg, func(o *bedrockruntime.Options) {
		o.Region = AWS_REGION
	})

	// Optional: Define inference parameters
	inferenceConfig := &bedrockTypes.InferenceConfiguration{
		MaxTokens: aws.Int32(500), // Limit response length
	}

	// Create converse request for Messages API
	converseRequest := &bedrockruntime.ConverseInput{
		ModelId: aws.String(MODEL_ID),
		Messages: []bedrockTypes.Message{{Role: bedrockTypes.ConversationRoleUser, Content: []bedrockTypes.ContentBlock{
			&bedrockTypes.ContentBlockMemberText{Value: prompt},
		}}},
		System: []bedrockTypes.SystemContentBlock{
			&bedrockTypes.SystemContentBlockMemberText{Value: "You are a helpful AI assistant that specializes in movie summaries in 100 words. Just return the summary."},
		},
		InferenceConfig: inferenceConfig,
	}

	output, err := bedrockRuntimeClient.Converse(context.TODO(), converseRequest)
	if err != nil {
		fmt.Println(err)
	}

	// Extract and print the response
	responseMessage := output.Output.(*bedrockTypes.ConverseOutputMemberMessage).Value
	// for _, content := range responseMessage.Content {
	// 	if text, ok := content.(*bedrockTypes.ContentBlockMemberText); ok {
	// 		fmt.Printf("Response: %s\n", text.Value)
	// 	}
	// }

	if len(responseMessage.Content) != 0 {
		response := responseMessage.Content[0].(*bedrockTypes.ContentBlockMemberText).Value
		if response != "" {
			fmt.Println(response)
		}
	}

	// Process the response
	// if output.Output == nil || output.Output.Message == nil || output.Output.Message.Content == nil {
	// 	return errors.New("received empty response from Bedrock API")
	// }
	// output.Output
	// // Print the response content
	// fmt.Println("Response:", *output.Output.Message.Content)

	return nil
}

func UpdateMovie(movieId string, year string, summary string) error {
	fmt.Println("UpdateMovie_DB")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	updateExpr := expression.Set(expression.Name("generatedSummary"), expression.Value(summary))
	expr, err := expression.NewBuilder().WithUpdate(updateExpr).Build()

	if err != nil {
		fmt.Printf("Couldn't build expression for update. Here's why: %v\n", err)
	} else {
		result, err := dynamoDb_Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
			TableName: aws.String(TABLE_NAME),
			Key: map[string]types.AttributeValue{
				"movieId":     &types.AttributeValueMemberS{Value: movieId},
				"releaseYear": &types.AttributeValueMemberN{Value: year},
			},
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ReturnValues:              types.ReturnValueAllNew,
		})

		if err != nil {
			fmt.Println(err)
		}
		var movie Movie
		if err := attributevalue.UnmarshalMap(result.Attributes, &movie); err != nil {
			fmt.Printf("Couldn't unmarshall update response. Here's why: %v\n", err)
		}
		fmt.Println(movie)
	}

	return nil
}

func GetMovieById(movieId string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	result, err := dynamoDb_Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"movieId": &types.AttributeValueMemberS{
				Value: movieId,
			},
		},
	},
	)

	if err != nil {
		fmt.Println(err)
	}
	if len(result.Item) == 0 {
		fmt.Println("no movie found")
	}
	var movie Movie
	if err := attributevalue.UnmarshalMap(result.Item, &movie); err != nil {
		fmt.Println(err)
	}

	fmt.Println(movie)
	return nil
}

func DeleteMovieById(movieId string) error {
	log.Print("Inside DeleteMovieById_DB func")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	res, err := dynamoDb_Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"movieId": &types.AttributeValueMemberS{
				Value: movieId,
			},
		},
		ConditionExpression: aws.String("attribute_exists(movieId)"),
	})

	fmt.Print(res)

	if err != nil {
		var cfe *types.ConditionalCheckFailedException
		if errors.As(err, &cfe) {
			fmt.Println("Conditional check failed:", cfe.Error())
		} else {
			fmt.Println("PutItem error:", err)
		}
	} else {
		fmt.Println("PutItem succeeded!")
	}

	if err != nil {
		fmt.Printf("failed to delete item from DynamoDB: %v\n", err)
		return nil
	}

	return nil
}

func AddMovie(movie Movie) error {
	fmt.Println("AddMovie")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	// check if movie is being created with same title
	res, _ := GetMovieByTitle_DB(movie.Title)

	if strings.Trim(strings.ToLower(res.Title), " ") == strings.Trim(strings.ToLower(movie.Title), " ") {
		log.Printf("Movie with same title already exists")
		return err
	}

	item, err := attributevalue.MarshalMap(movie)

	if err != nil {
		log.Printf("Couldn't marshall response. Here's why: %v\n", err)
		return err
	}

	result, err := dynamoDb_Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(TABLE_NAME),
		Item:      item,
	})

	if err != nil {
		log.Printf("Couldn't add item to table. Here's why: %v\n", err)
		return err
	}

	var movie1 Movie
	if err := attributevalue.UnmarshalMap(result.Attributes, &movie1); err != nil {
		log.Printf("Couldn't unmarshal reponse, %v", err)
		return err
	}

	log.Print(movie1)

	return nil
}

func GetMovieByTitle_DB(title string) (Movie, error) {
	log.Print("Inside GetMoviesByTitle_DB func")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
		return Movie{}, err
	}

	dynamoDb_Client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.Region = AWS_REGION
	})

	keyEx := expression.Key("title").Equal(expression.Value(title))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()

	if err != nil {
		return Movie{}, err
	}

	result, err := dynamoDb_Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:                 aws.String(TABLE_NAME),
		FilterExpression:          expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	log.Print(result)
	if err != nil {
		return Movie{}, err
	}
	var movie Movie
	if err := attributevalue.UnmarshalMap(result.Items[0], &movie); err != nil {
		return Movie{}, err
	}
	log.Print(movie)
	return movie, nil
}

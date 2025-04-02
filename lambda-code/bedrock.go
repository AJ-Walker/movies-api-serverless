package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

var BedrockClient *bedrockruntime.Client

func Init_Bedrock() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AWS_REGION))
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}
	BedrockClient = bedrockruntime.NewFromConfig(cfg)
}

func GenerateMovieSummary(movie Movie) (string, error) {
	log.Print("Inside GenerateMovieSummary func")
	// Define inference parameters
	inferenceConfig := &types.InferenceConfiguration{
		MaxTokens: aws.Int32(500), // Limit response length
	}

	prompt := fmt.Sprintf("Provide a short summary of 100 words for the movie '%v', released in %d, which falls under the genre %v.", movie.Title, movie.ReleaseYear, movie.Genre)

	// Create converse request for Messages API
	converseRequest := &bedrockruntime.ConverseInput{
		ModelId: aws.String(MODEL_ID),
		Messages: []types.Message{{Role: types.ConversationRoleUser, Content: []types.ContentBlock{
			&types.ContentBlockMemberText{Value: prompt},
		}}},
		System: []types.SystemContentBlock{
			&types.SystemContentBlockMemberText{Value: "You are a helpful AI assistant that specializes in movie summaries in 100 words. Just return the summary."},
		},
		InferenceConfig: inferenceConfig,
	}

	output, err := BedrockClient.Converse(context.TODO(), converseRequest)
	if err != nil {
		log.Print(err)
		return "", err
	}

	outputValue := output.Output.(*types.ConverseOutputMemberMessage).Value
	if len(outputValue.Content) == 0 {
		return "", fmt.Errorf("no summary returned")

	}
	result := outputValue.Content[0].(*types.ContentBlockMemberText).Value
	if result == "" {
		return "", fmt.Errorf("no summary returned")
	}

	return result, nil
}

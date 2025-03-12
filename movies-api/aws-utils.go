package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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

// func PutObject_S3(movies []Movie) {

// }

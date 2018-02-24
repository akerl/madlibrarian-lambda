package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/akerl/madlibrarian/utils"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bucketName := getParam(request, "bucket")
	storyName := getParam(request, "story")

	if bucketName == "" || storyName == "" {
		return fail("settings not provided")
	}

	config, err := configDownload(bucketName, storyName)
	if err != nil {
		return fail("config not found")
	}

	story, err := utils.NewStoryFromText(config)
	if err != nil {
		return fail("failed to parse config")
	}
	quote, err := story.Generate()
	if err != nil {
		return fail("failed to generate quote")
	}

	return events.APIGatewayProxyResponse{
		Body:       quote,
		StatusCode: 200,
	}, nil
}

func configDownload(bucketName, storyName string) ([]byte, error) {
	awsConfig := aws.NewConfig().WithCredentialsChainVerboseErrors(true)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := s3.New(sess)
	obj, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fmt.Sprintf("meta/%s.yml", storyName)),
	})
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(obj.Body)
}

func getParam(request events.APIGatewayProxyRequest, name string) string {
	res := request.PathParameters[name]
	if res == "" {
		res = os.Getenv(name)
	}
	return res
}

func fail(msg string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       msg,
		StatusCode: 500,
	}, nil
}

func main() {
	lambda.Start(Handler)
}

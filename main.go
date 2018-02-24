package main

import (
	"encoding/json"
	"fmt"

	"github.com/akerl/go-lambda/apigw"
	"github.com/akerl/go-lambda/s3"
	"github.com/akerl/madlibrarian/utils"
)

func loadQuote(params apigw.Params) (string, error) {
	bucketName := params.Lookup("bucket")
	storyName := params.Lookup("story")
	storyObject := fmt.Sprintf("meta/%s.yml", storyName)

	if bucketName == "" || storyName == "" {
		return "", fmt.Errorf("settings not provided")
	}

	config, err := s3.GetObject(bucketName, storyObject)
	if err != nil {
		return "", fmt.Errorf("config not found")
	}

	story, err := utils.NewStoryFromText(config)
	if err != nil {
		return "", fmt.Errorf("failed to parse config")
	}
	quote, err := story.Generate()
	if err != nil {
		return "", fmt.Errorf("failed to generate quote")
	}

	return quote, nil
}

type slackMessage struct {
	Text         string `json:"text,omitempty"`
	ResponseType string `json:"response_type,omitempty"`
}

func defaultHandler(req apigw.Request, params apigw.Params) (string, error) {
	return loadQuote(params)
}

func slackHandler(req apigw.Request, params apigw.Params) (string, error) {
	quote, err := loadQuote(params)
	if err != nil {
		return "", err
	}
	msg := &slackMessage{
		Text:         quote,
		ResponseType: "in_channel",
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response")
	}
	return string(jsonMsg), nil
}

func main() {
	lambda := apigw.Lambda{
		Handlers: map[string]apigw.Handler{
			"default": defaultHandler,
			"slack":   slackHandler,
		},
	}
	apigw.Start(lambda)
}

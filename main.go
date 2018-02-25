package main

import (
	"fmt"

	"github.com/akerl/go-lambda/apigw"
	"github.com/akerl/go-lambda/s3"
	"github.com/akerl/madlibrarian/utils"
)

func loadQuote(req apigw.Request, params apigw.Params) (interface{}, error) {
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

func main() {
	r := apigw.Router{
		Handlers: apigw.HandlerSet{
			&apigw.SlackHandler{
				Func: loadQuote,
			},
			&apigw.TextHandler{
				Func: loadQuote,
			},
		},
	}
	apigw.Start(r)
}

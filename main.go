package main

import (
	"fmt"

	"github.com/akerl/go-lambda/apigw"
	"github.com/akerl/go-lambda/s3"
	"github.com/akerl/madlibrarian/utils"
)

type storySet map[string]*utils.Story
type bucketSet map[string]storySet

var cache = make(bucketSet)

func cacheStory(bucketName, storyName string) (*utils.Story, error) {
	if cache[bucketName] == nil {
		cache[bucketName] = make(storySet)
	}
	if cache[bucketName][storyName] == nil {
		storyObject := fmt.Sprintf("meta/%s.yml", storyName)

		config, err := s3.GetObject(bucketName, storyObject)
		if err != nil {
			return &utils.Story{}, fmt.Errorf("config not found")
		}

		story, err := utils.NewStoryFromText(config)
		if err != nil {
			return &utils.Story{}, fmt.Errorf("failed to parse config")
		}
		cache[bucketName][storyName] = &story
	}
	return cache[bucketName][storyName], nil
}

func loadQuote(req apigw.Request, params apigw.Params) (interface{}, error) {
	bucketName := params.Lookup("bucket")
	storyName := params.Lookup("story")

	if bucketName == "" || storyName == "" {
		return "", fmt.Errorf("settings not provided")
	}

	story, err := cacheStory(bucketName, storyName)
	if err != nil {
		return "", err
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

package main

import (
	"context"
	"fmt"

	"github.com/akerl/madlibrarian/utils"
	"github.com/aws/aws-lambda-go/lambda"
)

// Request describes an incoming event"
type Request interface{}

// Response describes the result of processing the event
type Response interface{}

var stories map[string]utils.Story

func init() {
}

func main() {
	lambda.Start(Handler)
}

// Handler describes how to respond to API GW requests
func Handler(ctx context.Context, req Request) (Response, error) {
	fmt.Printf("%#v\n%#v\n", ctx, req)
	resp := "foo"
	return resp, nil
}

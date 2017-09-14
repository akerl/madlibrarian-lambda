package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/akerl/madlibrarian/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
)

type apigwEvent struct {
	PathParameters struct {
		Story string
	}
	StageVariables struct {
		Bucket string
	}
}

// Handle is the exported handler called by AWS Lambda.
var Handle apigatewayproxy.Handler

func init() {
	ln := net.Listen()
	Handle = apigatewayproxy.New(ln, []string{}).Handle
	go http.Serve(ln, http.HandlerFunc(handle))
}

func handle(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("X-Apigatewayproxy-Event")
	if header == "" {
		fail(w, "Not called from APIGW")
	}
	event := apigwEvent{}
	err := json.Unmarshal([]byte(header), &event)
	if err != nil {
		fail(w, "Header JSON deserialization failed")
	}
	storyName := event.PathParameters.Story
	bucket := event.StageVariables.Bucket
	if storyName == "" || bucket == "" {
		fail(w, "Variables not provided")
	}

	config, err := configDownload(bucket, storyName)
	if err != nil {
		fail(w, "Config not found")
	}

	s, err := utils.NewStoryFromText(config)
	if err != nil {
		fail(w, "Failed to parse config")
	}
	q, err := s.Generate()
	if err != nil {
		fail(w, "Failed to generate quote")
	}
	write(w, q)
}

func fail(w http.ResponseWriter, s string) {
	w.WriteHeader(http.StatusInternalServerError)
	write(w, s)
	panic()
}

func write(w http.ResponseWriter, s string) {
	w.Write([]byte(s))
}

func configDownload(bucket, storyName string) (string, error) {
	awsConfig := aws.NewConfig().WithCredentialsChainVerboseErrors(true)
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	}))
	client = s3.New(sess)
	obj, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("meta/%s.yml", storyName)),
	})
	if err != nil {
		return "", err
	}
	return ioutil.ReadAll(maxObj.Body)
}

func main() {}

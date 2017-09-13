package main

import (
	"fmt"
	"net/http"

	"github.com/akerl/madlibrarian/utils"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
)

// Handle is the exported handler called by AWS Lambda.
var Handle apigatewayproxy.Handler

func init() {
	ln := net.Listen()
	Handle = apigatewayproxy.New(ln, []string{}).Handle
	go http.Serve(ln, http.HandlerFunc(handle))
}

func handle(w http.ResponseWriter, r *http.Request) {
	s, err := utils.NewStoryFromPath("https://gist.githubusercontent.com/akerl/9321889d817beaddae2b66323e6b5a18/raw/1076def51a8d9beebddce075a0d841d2197c3bfb/gistfile1.txt")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("%v", err)))
	}
	q, err := s.Generate()
	if err != nil {
		w.Write([]byte(fmt.Sprintf("%v", err)))
	}
	w.Write([]byte(q))
}

func main() {}

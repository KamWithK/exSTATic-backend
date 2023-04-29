package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest() (string, error) {
	return "Test", nil
}

func main() {
	lambda.Start(HandleRequest)
}

package main

import (
	"github.com/1414C/cwl/handler"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(cwl.EC2ListCmd)
}

# cwl

## Overview

A set of AWS Lambda functions are contained in the m*n* folders.  The goal is to experiment / demonstrate the use of the AWS EC2 SDK from within the Lambda and Step-Function environments.

Each m*n* folder contains a handler coded in go.  Implementation of each handler's processing logic is contained in the ../handler/handlers.go file.  Each m*n* folder (package) compiles its own main function in its own main package.  This is an AWS requirement(?) and is the reason for the somewhat unorthodox project layout.  To clarify; the as-is project layout was chosen to permit the grouping of AWS Lambda functions in a single project based on area-of-use/purpose.

## Creating a Lambda function

1. We will code a new AWS Lambda function to read the status of one or more EC2 instances and report them to stdout.

2. Create a new folder in the project.  m*n* is the format to-date, but any name can be used.  For the purposes of this walkthrough, we will create folder m5.

3. Create new source file getec2statuses.go in the m5 folder as shown below.  This file will contain source-code used to build the handler for the new Lambda function.  If you are using an IDE with go tooling installed you will see complaints about cwl.GetEC2Statuses not existing.  This can be ignored for now.

```golang

    package main

    import (
      "github.com/1414C/cwl/handler"
      "github.com/aws/aws-lambda-go/lambda"
    )

    func main() {
      lambda.Start(cwl.GetEC2Statuses)
    }

```

4. Open the handlers.go source code file and add the GetEC2Statuses function defintion as shown below:

```golang

    // GetEC2Statuses is a test function for Lambda->EC2 AWS SDK access,
    // the purpose of which is to write the statuses of the selected EC2
    // instances to stdout.
    func GetEC2Statuses(event GetEC2InstancesEvent2) (string, error) {

      return "", nil
    }

```

5. Recall the Lambda functions are triggered via AWS Events, and the Events can be generated from many sources in the AWS environment.  The standard interface for a Lambda function allows for an incoming event of *the specified type* to be passed into the function at the time that it is called.  Our new function accepts the *GetEC2InstancesEvent2*, which has already been declared as follows in the handlers.go file:

```golang

    // GetEC2InstancesEvent2 is a test event structure for Lambda->EC2 access.
    type GetEC2InstancesEvent2 struct {
      Instances []string `json:"instances"`
    }

```

The event structure contains a single element which is intended to hold a slice of EC2 instance names.  If a slice is not provided in the event, the Instances element will contain nil and we will default the function to return the status of all instances in the target region.

6. Next we will add some logic to the new function defintion to illustrate how to access the incoming event element.  Update the function definition of GetEC2Statuses as follows:

```golang

    // GetEC2Statuses is a test function for Lambda->EC2 AWS SDK access,
    // the purpose of which is to write the statuses of the selected EC2
    // instances to stdout.
    func GetEC2Statuses(event GetEC2InstancesEvent2) (string, error) {

        // this writes to stdout, but does not update the AWS CloudWatch
        // log stream
        fmt.Println("loading function...")

        // log the received event, this will write the raw event to the
        // CloudWatch log stream
        log.Println("received event:", event)

        return "", nil
    }

```

7. For the purposes of this example, the AWS Region will be hard-coded as 'us-west-2', but it would be possible to provide this via the event structure.  The Lambda function will be executed with the IAM role that we specify in the aws create-function CLI call included in our build/deploy script.  Add the code to 'login' to the 'us-west-2' Region and create a new ec2 service as shown below:

```golang

    // GetEC2Statuses is a test function for Lambda->EC2 AWS SDK access,
    // the purpose of which is to write the statuses of the selected EC2
    // instances to stdout.
    func GetEC2Statuses(event GetEC2InstancesEvent2) (string, error) {

        // this writes to stdout, but does not update the AWS CloudWatch
        // log stream
        fmt.Println("loading function...")

        // log the received event, this will write the raw event to the
        // CloudWatch log stream
        log.Println("received event:", event)

        // using the IAM credentials asigned to the Lambda function, establish
        // a session in the 'us-west-2' AWS Region.  If a session cannot be
        // established, return an empty string and the error returned by the
        // AWS SDK NewSession(...) method.
        sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
        if err != nil {
          return "", err
        }

        // write the raw session information to the AWS CloudWatch stream
        fmt.Println("sess:", sess)

        // create a new instance of the EC2 client using the 'us-west-2' session
        // and return a simple error if the client creation fails.
        svc := ec2.New(sess)
        if svc == nil {
          return "", fmt.Errorf("failed to create EC2 client for us-west-2 session. session.Config follows: %v", sess.Config)
        }

        return "", nil
    }


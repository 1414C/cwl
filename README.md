# cwl

## Overview

A set of AWS Lambda functions are contained in the m*n* folders.  The goal is to experiment / demonstrate the use of the AWS EC2 SDK from within the Lambda and Step-Function environments.

Each m*n* folder contains a handler coded in go.  Implementation of each handler's processing logic is contained in the ../handler/handlers.go file.  Each m*n* folder (package) compiles its own main function in its own main package.  This is an AWS requirement(?) and is the reason for the somewhat unorthodox project layout.  To clarify; the as-is project layout was chosen to permit the grouping of AWS Lambda functions in a single project based on area-of-use/purpose.

## Creating a Lambda function in Go

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

The event structure contains a single element which is intended to hold a slice of EC2 instance names.  If a slice is not provided in the event, the Instances element will contain *nil* and we will default the function to return the status of all instances in the target region.

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

```

8. The next part does the actual job of taking the event input, formatting it and then using it to make a call using the AWS SDK ec2 client.  The ec2 client is comprehensive and directly exposes a method that can be used to obtain the status information from a set of instances in the targeted AWS Region.  The ec2 method accepts a slice of ec2 instance names, but can also be called with a *nil* value.  Calling with a *nil* results in the method returning status information for every instance in the caller's region.  Look closely at the input structures used by the various methods in the ec2 client API to ensure that the request matches up with your result expectations.

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
	svc := ec2.New(sess)
	if svc == nil {
		return "", fmt.Errorf("failed to create EC2 client for us-west-2 session. session.Config follows: %v", sess.Config)
	}

	// declare a variable to hold the result of the AWS SDK call to
	// ec2.DescribeInstanceStatus(..)
	var result *ec2.DescribeInstanceStatusOutput

	// if no EC2 instance names were provided by the event, call the AWS
	// SDK ec2.DescribeInstanceStatuses method without an instance list
	// and return the result.  Otherwise, iterate through the slice of
	// EC2 instances provided in the incoming event and build a slice of
	// string pointers as required be the the AWS SDK ec2.DescribeInstanceStatusInput
	// struct.  Next, call the ec2.DescribeInstanceStatuses method with the
	// input structure to get the statuses of the EC2 instances.  Errors
	// will be returned to the caller (AWS Lambda runtime).
	if event.Instances == nil {
		result, err = svc.DescribeInstanceStatus(nil)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
	} else {
		// populate a ec2.DescribeInstanceStatusInput struct based on
		// the instance-id's.
		var instIds []*string
		for _, inst := range event.Instances {
			instIds = append(instIds, aws.String(inst))
		}

		input := &ec2.DescribeInstanceStatusInput{
			InstanceIds:         instIds,
			IncludeAllInstances: aws.Bool(true),  // include stopped/terminated instances
			DryRun:              aws.Bool(false), // convert to *
		}

		result, err = svc.DescribeInstanceStatus(input)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
	}

	// no error, but no instances were found
	if result == nil || result.InstanceStatuses == nil {
		return "", nil
	}

	// write the instance statuses to stdout
	for _, v := range result.InstanceStatuses {
		fmt.Printf("instance-id: %s, instance-state: %s, instance-status: %s, system-status: %s\n", *v.InstanceId, *v.InstanceState, *v.InstanceStatus, *v.SystemStatus.Status)
	}

	// fmt.Println("Success", result)
	return result.String(), nil
}

```

9. At this point, the method is functionally complete, but does not pass the result back to the caller in an organized manner.  Calling the Stringer on the result is a handy way of veryifying the method is able to return what you are looking for, but is not what we should do long-term.  Note that we do not provide the result in JSON format, but as a go struct-type with json tags.  The runtime will marshal our return-type into JSON for consumption by the caller.  For now, we will leave the return parameter as-is and focus on getting the new function into AWS Lambda so we can test it.

## Adding a new go function to AWS Lambda

Once a new function has been declared and it's handler implemented, the next step is building and deploying the function in AWS Lambda.  The steps that are required are as follows:

- Make sure that the aws CLI tools have been installed on your local machine.
- Follow the CLI setup recommendations/procedures and ensure that your profile information is correct in the ~/.aws directory on your local machine.
- Set the AWS_PROFILE environment variable with the profile you wish to use for the push to AWS.
- Invoke the aws CLI tool to delete any existing function with the same name in AWS Lambda.
- Build the new handler for the target run environment.
- Update the permissions of the new executable.
- Compress the executable.
- Invoke the aws CLI tool to upload the zipped executable to AWS and create the new Lambda function.

### Steps

1. Open a terminal window and verify that you can call the aws tool.  You should see something like this:

```bash

$ which aws
/usr/local/bin/aws

```

If the *Which* command does not find the aws tool, check your $PATH and verify that the AWS CLI tools have been installed.
The CLI toolset and installation instructions are available at: <https://aws.amazon.com/cli/>

2. Set the AWS_PROFILE environment variable with the profile you wish to use to push / create the new Lambda function in AWS.  You may check your profiles in the ~/.aws/profile file in your $HOME directory.  If you do not see a *config* and *credentials* file in ~/.aws, go back to the AWS CLI instructions at <https://aws.amazon.com/cli/> and follow the steps to confgure your CLI environment.

Set and check the AWS_PROFILE environment variable as follows:

```bash

$ export AWS_PROFILE=myprofile
$ echo $AWS_PROFILE
myprofile

```

3. It is possible that there is an existing AWS Lambda function with the same name in your target environment.  There are a few options in this case, but for now we will simply invoke the *aws* CLI tool to delete any existing Lambda function sharing the same as our new function.

```bash

$ aws lambda delete-function --function-name GetEC2InstanceStatuses
An error occurred (ResourceNotFoundException) when calling the DeleteFunction operation: Function not found: arn:aws:lambda:us-west-2:907538708243:function:GetEC2InstanceStatuses

```

If the function did not exist, you will see an error message as shown above.  This is fine, and we will come back to this step when we modify the function to pass back a more meaningful result.

4. The new function must be compiled for the target operating system and CPU architecture before it can be pushed up to AWS Lambda.  In the case of Go this is quite easy, as it offers simple cross-compilation via the setting of two environment variables.  This demo was written on a mac running OS X, but our target EC2 environment is running EC2 Linux.  Both systems are running on amd64, but the GOARCH has been set for illustrative purposes.  A list of valid GOOS and GOARCH values can be found at <https://golang.org/doc/install/source#environment>.  Remember that you are specifying the *target* environement here.

```bash

$ GOOS=linux GOARCH=amd64 go build -o main getec2statuses.go
$ ls -l
total 33008
-rw-r--r--  1 stevem  staff       147 May 31 09:19 getec2statuses.go
-rwxr-xr-x  1 stevem  staff  16893531 May 31 11:56 main
$

```

5. If the permissions of the newly generated executable allow execution at all three levels of permission continue to step 6, otherwise run the following command to grant the required permissions:

```bash

$ chmod 555 main

```

6. Compress the *main* executable into a file called *deployment.zip*.

```bash

$ zip deployment.zip ./main
  adding: main (deflated 72%)
$

```

7. Invoke the aws CLI tool to upload the *deployment.zip* file to AWS and create create the new Lambda function.  The aws tool has a lambda command annex and makes use of the following parameter when pushing a new function to AWS:

#### create-function Parameters

- ***function-name*** - This is the name of the function that will be created in AWS Lambda.  Best practice is to align this with the function-name used in the Go source code, although this is not a requirement.
- ***memory*** - This sets the amount of memory allocated on the EC2 instance (MB) when running the Lambda function.
- ***role*** - This is the AWS IAM Role that will be assigned to the new Lambda function.  You need to ensure that the Role you specify here has enough access to allow the Lambda function to run, as well as any other Policies that are required.  In our case, the Role must contain write access to CloudWatch, as we make use of the log.Println/log.Printf methods in our new Lambda function.  We are also making use of the AWS SDK to access EC2, so we need to ensure that the role offers EC2 access as well.
- ***runtime*** - Specifies the runtime to be used when executing the Lambda function in the AWS environment.
- ***zip-file*** - Specifies the location of the compiled and zipped function on the local machine.
- ***handler*** - Specifies the name of the handler function in the compiled source code.

Execute the *aws* CLI tool as follows:

```bash

$ aws lambda create-function --region us-west-2 \
--function-name GetEC2Statuses \
--memory 128 \
--role arn:aws:iam::907538708243:role/LambdaEC2Access \
--runtime go1.x \
--zip-file fileb:///Users/stevem/gowork/src/github.com/1414C/cwl/m4/deployment.zip \
--handler main

{
    "TracingConfig": {
        "Mode": "PassThrough"
    },
    "CodeSha256": "dTgIO9Vhg3RBJ3fqSIxyCgYOX7xsDHx0s3iSMQ15KLQ=",
    "FunctionName": "GetEC2Statuses",
    "CodeSize": 4694387,
    "RevisionId": "8b1290da-94ce-44a5-8636-829490402268",
    "MemorySize": 128,
    "FunctionArn": "arn:aws:lambda:us-west-2:907538708243:function:GetEC2Statuses",
    "Version": "$LATEST",
    "Role": "arn:aws:iam::907538708243:role/LambdaEC2Access",
    "Timeout": 3,
    "LastModified": "2018-05-31T18:16:53.195+0000",
    "Handler": "main",
    "Runtime": "go1.x",
    "Description": ""
}
$

```

## Testing

To test the new Lambda function, login to the AWS console and select the Lambda Service from the drop-down services menu.  You should be able to select the new function from the list of Lambdas as shown below:

![Lambda Functions](https://github.com/1414C/cwl/raw/master/images/Lambda1.jpeg "Lambda Functions")


Click into the function and then select the 'Select a test event...' drop-down in the upper right corner of the screen and choose the 'Configure test event' option:

![Configure Test Event](https://github.com/1414C/cwl/raw/master/images/Lambda2.jpeg "Configure Test Event")

Setup a test event for a single EC2 instance in your account/region as shown below:

![Configure Single Instance Test Event](https://github.com/1414C/cwl/raw/master/images/Lambda3.jpeg "Configure Test Event")

While we are here, setup another test event with more than one EC2 instance in the test event instances element:

![Configure Multiple Instance Test Event](https://github.com/1414C/cwl/raw/master/images/Lambda4.jpeg "Configure Test Event")

Next, we will trigger the Lambda function using via the test tool making use of the new single-instance test event:

![Trigger Single Instance Test Event](https://github.com/1414C/cwl/raw/master/images/Lambda5.jpeg "Trigger Single Instance Test Event")
We can see that the function ran successfully and wrote a Stringified version of the function's result variable to stdout.  Notice that a CloudWatch log has also been generated containing details of the function's execution.  In this case, the instance being queried was in a stopped state.  We will start the instance via the EC2 dashboard and test the function again.

![Testing Instance Pending](https://github.com/1414C/cwl/raw/master/images/Lambda6.jpeg "Testing Instance Pending")
A start has been requested for the EC2 instance and we can see that the instance state is now 'pending'.

![Testing Instance Running](https://github.com/1414C/cwl/raw/master/images/Lambda6.jpeg "Testing Instance Running")
After a brief wait, the test event was triggered again and we can see that the test instance now has an instance-state of 'running'.  This does not indicate that the instance is ready for business, but rather that the image has been started.  Notice that the 'Reachability' of the instance is shown as 'initiallizing'.  This status corresponds to the 'Status Checks' column in the EC2 instance overview in the EC2 dashboard.

![Testing Instance Running](https://github.com/1414C/cwl/raw/master/images/Lambda6.jpeg "Testing Instance Running")
After another brief wait, the test event was triggered again and we can see that the test instance has a instance-state of 'running' and the 'reachability' and 'system-status' are reported as 'passed' and 'ok' respectively.  The instance appears to be up and running.

From here, we could excute other Lambda functions against the instance to do things like get its internal and external IP addresses, execute commands and call services etc.

## Cleaning up the function return parameters

Returning a Stringified result is not terribly useful to the caller, so seeing as the EC2 API returns a fairly concise dataset we will simply return the []*ec2.InstanceStatus type.  A complete and updated version of the GetEC2Statuses function follows:

```golang

// GetEC2Statuses is a test function for Lambda->EC2 AWS SDK access,
// the purpose of which is to write the statuses of the selected EC2
// instances to stdout.
func GetEC2Statuses(event GetEC2StatusesEvent) ([]*ec2.InstanceStatus, error) {

	// this writes to stdout, and updates the AWS CloudWatch
	// log stream
	log.Println("loading function...")

	// log the received event, this will write the raw event to the
	// CloudWatch log stream
	log.Println("received event:", event.Instances)

	// using the IAM credentials asigned to the Lambda function, establish
	// a session in the 'us-west-2' AWS Region.  If a session cannot be
	// established, return an empty string and the error returned by the
	// AWS SDK NewSession(...) method.
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		return nil, err
	}

	// create a new instance of the EC2 client using the 'us-west-2' session
	svc := ec2.New(sess)
	if svc == nil {
		return nil, fmt.Errorf("failed to create EC2 client for us-west-2 session. session.Config follows: %v", sess.Config)
	}

	// declare a variable to hold the result of the AWS SDK call to
	// ec2.DescribeInstanceStatus(..)
	var result *ec2.DescribeInstanceStatusOutput

	// if no EC2 instance names were provided by the event, call the AWS
	// SDK ec2.DescribeInstanceStatuses method without an instance list
	// and return the result.  Otherwise, iterate through the slice of
	// EC2 instances provided in the incoming event and build a slice of
	// string pointers as required be the the AWS SDK ec2.DescribeInstanceStatusInput
	// struct.  Next, call the ec2.DescribeInstanceStatuses method with the
	// input structure to get the statuses of the EC2 instances.  Errors
	// will be returned to the caller (AWS Lambda runtime).
	if event.Instances == nil {
		result, err = svc.DescribeInstanceStatus(nil)
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
	} else {
		// populate a ec2.DescribeInstanceStatusInput struct based on
		// the instance-id's.
		var instIds []*string
		for _, inst := range event.Instances {
			instIds = append(instIds, aws.String(inst))
		}

		input := &ec2.DescribeInstanceStatusInput{
			InstanceIds:         instIds,
			IncludeAllInstances: aws.Bool(true),  // include stopped/terminated instances
			DryRun:              aws.Bool(false), // convert to *
		}

		result, err = svc.DescribeInstanceStatus(input)
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
	}

	// no error, but no instances were found
	if result == nil || result.InstanceStatuses == nil {
		return nil, nil
	}

	// write the instance statuses to stdout
	for _, v := range result.InstanceStatuses {
		log.Printf("instance-id: %s, instance-state: %s, instance-status: %s, system-status: %s\n", *v.InstanceId, *v.InstanceState, *v.InstanceStatus, *v.SystemStatus.Status)
		log.Printf("system-status: %s\n", *v.SystemStatus.Status)
		for _, d := range v.SystemStatus.Details {
			log.Printf("system-status name %v impaired since %v\n", d.Name, d.ImpairedSince)
		}
	}
	return result.InstanceStatuses, nil
}

```

A call of the completed GetEC2Statuses function returns the following to the caller:

```JSON

[
  {
    "AvailabilityZone": "us-west-2a",
    "Events": null,
    "InstanceId": "i-02b6e4d7e690a090d",
    "InstanceState": {
      "Code": 80,
      "Name": "stopped"
    },
    "InstanceStatus": {
      "Details": null,
      "Status": "not-applicable"
    },
    "SystemStatus": {
      "Details": null,
      "Status": "not-applicable"
    }
  },
  {
    "AvailabilityZone": "us-west-2a",
    "Events": null,
    "InstanceId": "i-055352fea72019e3f",
    "InstanceState": {
      "Code": 16,
      "Name": "running"
    },
    "InstanceStatus": {
      "Details": [
        {
          "ImpairedSince": null,
          "Name": "reachability",
          "Status": "passed"
        }
      ],
      "Status": "ok"
    },
    "SystemStatus": {
      "Details": [
        {
          "ImpairedSince": null,
          "Name": "reachability",
          "Status": "passed"
        }
      ],
      "Status": "ok"
    }
  }
]

```
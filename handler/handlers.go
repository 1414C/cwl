package cwl

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	// "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/batch"
	"log"
)

type JobEvent struct {
	JobName       string `json:"jobName"`
	JobDefinition string `json:"jobDefinition"`
	JobQueue      string `json:"jobQueue"`
	WaitTime      int    `json:"wait_time"`
}

// JobGuid is the event input structure containing the
// one-and-only input parameter for this function.
type JobGuid struct {
	JobID string `json:"jobID"`
}

// CheckJobFunc3 checks and returns the status of the job
// identified by event.JobID.  If a technical error is
// encountered an empty string and non-nil error are
// returned.
// The return string parameter is mapped to:
// "ResultPath": "%.status" in the State Machine Definition.
func CheckJobFunc3(event JobGuid) (string, error) {

	fmt.Println("loading function...")

	// log the received event
	log.Println("received event:", event)

	// create a new batch-session
	svc := batch.New(session.New())

	// setup the input values
	input := &batch.DescribeJobsInput{
		Jobs: []*string{
			aws.String(event.JobID),
		},
	}

	// get the job status
	result, err := svc.DescribeJobs(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case batch.ErrCodeClientException:
				fmt.Println(batch.ErrCodeClientException, aerr.Error())
			case batch.ErrCodeServerException:
				fmt.Println(batch.ErrCodeServerException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "", err
	}

	fmt.Println("result:", result)

	// return the job status; JobStatusFailed if no Job found for JobId
	if len(result.Jobs) > 0 {
		jobDetail := result.Jobs[0]

		// return response, nil
		return *jobDetail.Status, nil
	}

	// "status": <status>
	return batch.JobStatusFailed, nil
}

// SubmitJobFunc3 submits a job to AWS Batch based on the incoming
// event structure.  The Job must be defined in the AWS batch
// console (for this one anyway), and is part of the event
// struct.
// example input:
// {
//	  "jobName": "my-test-job-4d",
//	  "jobDefinition": "arn:aws:batch:us-west-2:755561232688:job-definition/SampleJobDefinition-e3e85ee22b798f7:1",
//	  "jobQueue": "arn:aws:batch:us-west-2:755561232688:job-queue/SampleJobQueue-40e2ee4d7b7d43b",
//	  "wait_time": 60
// }
func SubmitJobFunc3(event JobEvent) (JobGuid, error) {

	fmt.Println("loading function...")

	// log the received event
	log.Println("received event:", event)

	// create a new batch-session
	svc := batch.New(session.New())

	// setup the job submission parameters
	input := &batch.SubmitJobInput{
		JobDefinition: &event.JobDefinition,
		JobName:       &event.JobName,
		JobQueue:      &event.JobQueue,
	}

	// submit the job and then check for errors
	result, err := svc.SubmitJob(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case batch.ErrCodeClientException:
				fmt.Println(batch.ErrCodeClientException, aerr.Error())
			case batch.ErrCodeServerException:
				fmt.Println(batch.ErrCodeServerException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return JobGuid{}, err
	}

	fmt.Println("result:", result)

	// create the response in the format of:
	// "guid": {
	//		"jobID": "643685f9-26fa-4d8b-a291-aa2bbb925f76"
	// }
	// the "guid" response is mapped to:
	// "ResultPath": "$.guid"
	// in the "Submit Job" Task in the State Machine Definition
	response := JobGuid{
		JobID: *result.JobId,
	}
	return response, nil
}

// GetEC2InstancesEvent is a test event structure for Lambda->EC2 access.
type GetEC2InstancesEvent struct {
	Instance string `json:"instance"`
}

// GetEC2Instances is a test method for Lambda->EC2 AWS SDK access
func GetEC2Instances(event GetEC2InstancesEvent) (string, error) {

	fmt.Println("loading function...")

	// log the received event
	log.Println("received event:", event)

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		panic(err)
	}

	fmt.Println("sess:", sess)
	svc := ec2.New(sess)

	if event.Instance == "" {
		result, err := svc.DescribeInstances(nil)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
		fmt.Println("Success", result)
		return result.String(), nil
	} else {
		var instIds []*string
		instIds = append(instIds, aws.String(event.Instance))
		input := &ec2.DescribeInstancesInput{
			InstanceIds: instIds,
			DryRun:      aws.Bool(false), // convert to *
		}
		result, err := svc.DescribeInstances(input)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
		fmt.Println("Success", result)
		return result.String(), nil
	}
}

// GetEC2InstancesEvent2 is a test event structure for Lambda->EC2 access.
type GetEC2InstancesEvent2 struct {
	Instances []string `json:"instances"`
}

// GetEC2Instances2 is a test method for Lambda->EC2 AWS SDK access
func GetEC2Instances2(event GetEC2InstancesEvent2) (string, error) {

	fmt.Println("loading function...")

	// log the received event
	log.Println("received event:", event)

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		panic(err)
	}

	fmt.Println("sess:", sess)
	svc := ec2.New(sess)

	if event.Instances == nil {
		result, err := svc.DescribeInstances(nil)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		}
		fmt.Println("Success", result)
		return result.String(), nil
	}

	var instIds []*string
	for _, inst := range event.Instances {
		instIds = append(instIds, aws.String(inst))
	}

	log.Println("GOT:", instIds)

	input := &ec2.DescribeInstancesInput{
		InstanceIds: instIds,
		DryRun:      aws.Bool(false), // convert to *
	}
	result, err := svc.DescribeInstances(input)
	if err != nil {
		return "", fmt.Errorf("%s", err)
	}
	fmt.Printf("Got %d Reservations.\n", len(result.Reservations))
	if result.Reservations != nil {
		for _, v := range result.Reservations {
			if v.Instances != nil {
				fmt.Printf("Reservation %v has %d Instances:\n", v.ReservationId, len(v.Instances))
				for _, vi := range v.Instances {
					// fmt.Printf("instance-id: %s, instance-type: %s, instance-lifecycle: %s, launch-time: %v\n", *vi.InstanceId, *vi.InstanceType, *vi.InstanceLifecycle, vi.LaunchTime)
					fmt.Printf("instance-id: %s, instance-type: %s, launch-time: %v, public-ip: %s\n", *vi.InstanceId, *vi.InstanceType, *vi.LaunchTime, *vi.PublicIpAddress)
				}
			} else {
				fmt.Printf("Reservation %v has no Instances.\n", v.ReservationId)
			}
		}
	}
	// fmt.Println("Success", result)
	return result.String(), nil
}

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

	// if no EC2 instances were provided by the event, call the AWS SDK
	// ec2.DescribeInstanceStatuses method without an instance list and
	// return the result.  Otherwise, iterate through the slice of
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

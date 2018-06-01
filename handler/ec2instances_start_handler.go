package cwl

// smacleod - 2018-06-01
import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
)

// EC2InstancesStartEvent triggers function cwl.EC2InstancesStart.
type EC2InstancesStartEvent struct {
	Instances []string `json:"instances"`
}

// EC2InstancesStart is a test function, the purpose of which is to start the
// named EC2 Instances.  The function does not attempt to determine the status
// of the named instances prior to executing the start attempts.  Determination
// of instance status should be performed prior to calling this function.
func EC2InstancesStart(event EC2InstancesStartEvent) (*ec2.StartInstancesOutput, error) {

	// log the received event, this will write the raw event to the
	// CloudWatch log stream
	log.Println("loading function...")
	log.Println("received event:", event.Instances)

	// if no EC2 instance names were provided by the event, return an error.
	if event.Instances == nil {
		return nil, fmt.Errorf("no instance names were specified in triggering event %v", event)
	}

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
	// ec2.StartInstances(...)
	var result *ec2.StartInstancesOutput

	// Iterate through the slice of EC2 instances provided by the incoming
	// event and build a slice of string pointers as required be the AWS
	// SDK ec2.StartInstancesInput struct.
	// Next, call the ec2.StartInstances method with the input structure.
	// Errors / new system statuses will be returned to the caller.
	var instIds []*string
	for _, inst := range event.Instances {
		instIds = append(instIds, aws.String(inst))
	}

	input := &ec2.StartInstancesInput{
		AdditionalInfo: nil,
		InstanceIds:    instIds,
		DryRun:         aws.Bool(false), // convert to *
	}

	result, err = svc.StartInstances(input)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	// no error, also no result(possible?)
	if result == nil || result.StartingInstances == nil {
		return nil, fmt.Errorf("instance start for instances %v returned no information - status unknown", instIds)
	}

	return result, nil
}

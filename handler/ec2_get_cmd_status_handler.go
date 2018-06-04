package cwl

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
)

// EC2ListCmdEvent triggers function cwl.EC2ListCmd
type EC2ListCmdEvent struct {
	Cmd       string   `json:"cmd"`
	Instances []string `json:"instances"`
}

// EC2ListCmd lists the specified command status/properties
func EC2ListCmd(event EC2ListCmdEvent) (*ssm.ListCommandsOutput, error) {

	// log the received event, this will write the raw event to the
	// CloudWatch log stream
	log.Println("loading function...")
	log.Println("received event.Cmd:", event.Cmd)
	log.Println("received event.Instances:", event.Instances)

	// if no commandID was passed in the event, return an error.
	if event.Cmd == "" {
		return nil, fmt.Errorf("no command-id was provided in triggering event %v", event)
	}

	// using the IAM credentials asigned to the Lambda function, establish
	// a session in the 'us-west-2' AWS Region.  If a session cannot be
	// established, return an empty string and the error returned by the
	// AWS SDK NewSession(...) method.
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		return nil, err
	}

	// create a new instance of the SSM client using the 'us-west-2' session
	svc := ssm.New(sess)
	if svc == nil {
		return nil, fmt.Errorf("failed to create EC2 client for us-west-2 session. session.Config follows: %v", sess.Config)
	}

	listCommandsInput := ssm.ListCommandsInput{
		CommandId: aws.String(event.Cmd),
		// CommandId: result.Command.CommandId,
		// InstanceId: result.Command.InstanceIds[0],
		// MaxResults: aws.Int64(100),
	}

	listCommandsResult, err := svc.ListCommands(&listCommandsInput)
	if err != nil {
		fmt.Printf("error calling ssm.ListCommands for commandID: %s\n", event.Cmd)
		// Cast err to awserr.Error to handle specific error codes.
		aerr, ok := err.(awserr.Error)
		log.Printf("error code %s received for ssm.ListCommands call\n", aerr.Code())
		if ok && aerr.Code() == "400" {
			// Specific error code handling
		}
		return nil, err
	}
	log.Println(listCommandsResult)
	return listCommandsResult, nil
}

package cwl

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// EC2IssueCmdEvent triggers function cwl.EC2IssueCmd.
type EC2IssueCmdEvent struct {
	Instances []string `json:"instances"`
	Cmd       string   `json:"cmd"`
}

// EC2IssueCmd runs the specified command on the specified EC2 instances.
func EC2IssueCmd(event EC2IssueCmdEvent) (*ssm.Command, error) {

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

	// create a new instance of the SSM client using the 'us-west-2' session
	svc := ssm.New(sess)
	if svc == nil {
		return nil, fmt.Errorf("failed to create EC2 client for us-west-2 session. session.Config follows: %v", sess.Config)
	}

	// convert instanceIds to []*string
	var instIds []*string
	for _, inst := range event.Instances {
		instIds = append(instIds, aws.String(inst))
	}

	// build the command-map
	var params map[string][]*string
	var cmdStrings []*string

	// str1 := "mv /tmp/folder_two/test_file.txt /tmp/folder_one" // pwd"
	// cmdStrings = append(cmdStrings, &str1)
	cmdStrings = append(cmdStrings, &event.Cmd)
	params = make(map[string][]*string)
	params["commands"] = cmdStrings

	// targets can be specified in-place of instance-ids
	// for example, targets can be used to identify EC2 instances by tag, or
	// by fleet assignment.
	// var targets []*ssm.Target
	// target := &ssm.Target{
	// 	Key:    aws.String("instanceids"),
	// 	Values: instIds,
	// }
	// targets = append(targets, target)

	// setup the command
	commandInput := ssm.SendCommandInput{
		Comment:          aws.String("test comment"),
		DocumentHash:     nil,
		DocumentHashType: nil,
		DocumentName:     aws.String("AWS-RunShellScript"),
		InstanceIds:      instIds,
		MaxConcurrency:   aws.String("2"),
		MaxErrors:        aws.String("4"),
		//NotificationConfig: &ssm.NotificationConfig{  // setup SNS by EventType (success,failure,pending..)
		//	NotificationArn:    aws.String(""),
		//	NotificationEvents: nil,
		//	NotificationType:   aws.String(""),
		//},
		//OutputS3BucketName: aws.String(""),
		//OutputS3KeyPrefix:  aws.String(""),
		//OutputS3Region:     aws.String(""),
		Parameters: params,
		// ServiceRoleArn: aws.String(""),
		//Targets:        targets,
		TimeoutSeconds: aws.Int64(30), // minimum value = 30
	}

	result, err := svc.SendCommand(&commandInput)
	if err != nil {
		fmt.Println("error detected...")
		// Cast err to awserr.Error to handle specific error codes.
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == "400" {
			// Specific error code handling
		}
	}
	log.Println("SendCommandInput result:")
	log.Println(result)
	return result.Command, nil
}

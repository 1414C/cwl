export AWS_PROFILE=smacleod
aws lambda delete-function --function-name EC2InstancesStart
GOOS=linux go build -o main ec2instancestart.go
chmod 555 main
zip deployment.zip ./main
aws lambda create-function --region us-west-2 --function-name EC2InstancesStart --memory 128 --role arn:aws:iam::907538708243:role/LambdaEC2Access --runtime go1.x --zip-file fileb:///Users/stevem/gowork/src/github.com/1414C/cwl/m6/deployment.zip --handler main
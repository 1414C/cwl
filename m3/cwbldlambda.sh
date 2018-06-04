export AWS_PROFILE=smacleod
aws lambda delete-function --function-name GetEC2Instances
GOOS=linux go build -o main getec2instances.go
chmod 555 main
zip deployment.zip ./main
aws lambda create-function --region us-west-2 --function-name GetEC2Instances --memory 128 --role arn:aws:iam::907538708243:role/LambdaEC2Access --runtime go1.x --zip-file fileb:///Users/stevem/gowork/src/github.com/1414C/cwl/m3/deployment.zip --handler main

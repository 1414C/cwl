export AWS_PROFILE=smacleod
aws lambda delete-function --function-name CheckJobFunc3
GOOS=linux go build -o main checkfunc3.go 
zip deployment.zip ./main
aws lambda create-function --region us-west-2 --function-name CheckJobFunc3 --memory 128 --role arn:aws:iam::907538708243:role/SimpleJobSubmissionAndStatus --runtime go1.x --zip-file fileb:///Users/stevem/gowork/src/github.com/1414C/cwl/m1/deployment.zip --handler main
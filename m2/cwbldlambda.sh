export AWS_PROFILE=smacleod
aws lambda delete-function --function-name SubmitJobFunc3
GOOS=linux go build -o main submitfunc3.go
chmod 555 main
zip deployment.zip ./main
aws lambda create-function --region us-west-2 --function-name SubmitJobFunc3 --memory 128 --role arn:aws:iam::907538708243:role/SimpleJobSubmissionAndStatus --runtime go1.x --zip-file fileb:///Users/stevem/gowork/src/github.com/1414C/cwl/m2/deployment.zip --handler main
Set-Location $PSScriptRoot
if (Test-Path .\function.zip)
{
    Remove-Item .\function.zip
}
$env:GOARCH="amd64"
$env:GOOS="linux"
go build -o check check.go

Compress-Archive -Path .\check -DestinationPath .\function.zip
Remove-Item .\check

aws s3 cp .\function.zip s3://radswn-lambda-bucket
$FileVersion = (aws s3api list-object-versions --bucket radswn-lambda-bucket| ConvertFrom-Json).Versions[0].VersionId

aws cloudformation deploy --stack-name lambda-stack --template-file .\lambda.yaml --capabilities CAPABILITY_NAMED_IAM --parameter-overrides LambdaCodeVersion=$FileVersion
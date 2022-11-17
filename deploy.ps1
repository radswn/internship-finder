function Timestamp-Log($Text)
{
    $Timestamp = Get-Date -Format HH:mm:ss.fff
    $Text = $Text.toUpper()
    Write-Host "[$Timestamp] ============== $Text =============="
}

Set-Location $PSScriptRoot

Timestamp-Log("tests")
go test
if ($LASTEXITCODE -ne 0)
{
    Write-Output "There are test failures"
    exit 1
}

Timestamp-Log("build")
$env:GOARCH = "amd64"
$env:GOOS = "linux"
go build -o check check.go

Timestamp-Log("upload to bucket")
Compress-Archive -Path .\check -DestinationPath .\function.zip
Remove-Item .\check

aws s3 cp .\function.zip s3://radswn-lambda-bucket
Remove-Item .\function.zip

Timestamp-Log("stack update")
$FileVersion = (aws s3api list-object-versions --bucket radswn-lambda-bucket| ConvertFrom-Json).Versions[0].VersionId
aws cloudformation deploy --stack-name lambda-stack --template-file .\infra.yaml --capabilities CAPABILITY_NAMED_IAM --parameter-overrides LambdaCodeVersion = $FileVersion
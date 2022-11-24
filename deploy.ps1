function Timestamp-Log($Text)
{
    $Timestamp = Get-Date -Format HH:mm:ss.fff
    $Text = $Text.toUpper()
    Write-Host "[$Timestamp] ============== $Text =============="
}

$BucketName = "radswn-lambda-bucket"
$CheckerExe = "checker"
$CheckerLambdaZip = $CheckerExe + ".zip"
$CheckerSourceCode = $CheckerExe + ".go"

Set-Location $PSScriptRoot

$env:GOOS = "windows"
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
go build -o $CheckerExe $CheckerSourceCode

Timestamp-Log("upload to bucket")
Compress-Archive -Path .\$CheckerExe -DestinationPath .\$CheckerLambdaZip
Remove-Item .\$CheckerExe

aws s3 cp .\$CheckerLambdaZip s3://$BucketName
Remove-Item .\$CheckerLambdaZip

Timestamp-Log("stack update")
$CheckerFileVersion = ((aws s3api list-object-versions --bucket $BucketName | `
ConvertFrom-Json).Versions | Where-Object -Property Key -EQ $CheckerLambdaZip)[0].VersionId

aws cloudformation deploy --stack-name lambda-stack `
--template-file .\infra.yaml `
--capabilities CAPABILITY_NAMED_IAM `
--parameter-overrides CodeBucketName=$BucketName `
CheckerLambdaCodeVersion=$CheckerFileVersion `
CheckerLambdaZip=$CheckerLambdaZip
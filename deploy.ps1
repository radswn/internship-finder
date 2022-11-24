function Timestamp-Log($Text)
{
    $Timestamp = Get-Date -Format HH:mm:ss.fff
    $Text = $Text.toUpper()
    Write-Host "[$Timestamp] ============== $Text =============="
}

$BucketName = "radswn-lambda-bucket"

$CheckerExe = "checker\checker"
$CheckerLambdaZip = $CheckerExe + ".zip"
$CheckerSourceCode = $CheckerExe + ".go"


$NotifierExe = "notifier\notifier"
$NotifierLambdaZip = $NotifierExe + ".zip"
$NotifierSourceCode = $NotifierExe + ".go"


Set-Location $PSScriptRoot

$env:GOOS = "windows"
Timestamp-Log("tests")
go test .\checker
if ($LASTEXITCODE -ne 0)
{
    Write-Output "There are test failures"
    exit 1
}

Timestamp-Log("build")
$env:GOARCH = "amd64"
$env:GOOS = "linux"
go build -o $CheckerExe $CheckerSourceCode
go build -o $NotifierExe $NotifierSourceCode

Timestamp-Log("upload to bucket")
Compress-Archive -Path .\$CheckerExe -DestinationPath .\$CheckerLambdaZip
Compress-Archive -Path .\$NotifierExe -DestinationPath .\$NotifierLambdaZip

Remove-Item .\$CheckerExe
Remove-Item .\$NotifierExe

aws s3 cp .\$CheckerLambdaZip s3://$BucketName
aws s3 cp .\$NotifierLambdaZip s3://$BucketName

Remove-Item .\$CheckerLambdaZip
Remove-Item .\$NotifierLambdaZip

$CheckerLambdaZip = $CheckerLambdaZip.Split("\")[1]
$NotifierLambdaZip = $NotifierLambdaZip.Split("\")[1]

Timestamp-Log("stack update")
$CheckerFileVersion = ((aws s3api list-object-versions --bucket $BucketName | `
ConvertFrom-Json).Versions | Where-Object -Property Key -EQ $CheckerLambdaZip)[0].VersionId

$NotifierFileVersion = ((aws s3api list-object-versions --bucket $BucketName | `
ConvertFrom-Json).Versions | Where-Object -Property Key -EQ $NotifierLambdaZip)[0].VersionId

aws cloudformation deploy --stack-name lambda-stack `
--template-file .\infra.yaml `
--capabilities CAPABILITY_NAMED_IAM `
--parameter-overrides CodeBucketName=$BucketName `
CheckerLambdaCodeVersion=$CheckerFileVersion `
CheckerLambdaZip=$CheckerLambdaZip `
NotifierLambdaCodeVersion=$NotifierFileVersion `
NotifierLambdaZip=$NotifierLambdaZip

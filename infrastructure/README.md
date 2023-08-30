# exSTATic Backend

Creating a backend for exSTATic, hosted on AWS, created with the AWS CDK for TypeScript and Lambda's in Go

## Useful commands

* `npm run build`   compile typescript to js
* `npm run watch`   watch for changes and compile
<!-- * `npm run test`    perform the jest unit tests -->
* `cdk deploy`      deploy this stack to your default AWS account/region
* `cdk diff`        compare deployed stack with current state
* `cdk synth`       emits the synthesized CloudFormation template

## Local Testing
* `cd infrasturcture`
* `cdk synth`
* `cdklocal bootstrap`
* `cdklocal deploy -a "cdk.out/assembly-localStage/" --require-approval "never" --all`
* `cd ../functions`
* `go test ./...`

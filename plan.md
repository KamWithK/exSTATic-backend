# Costs
## Requests (24 hour support, 5 hour usage hour average, 1 request every 30 seconds, 100 users, 1 month)
<!-- ### Kinesis
0.049 * 24 * 30
+
(0.098 / 1000000) * (1 * (60 / 30) * 60 * 5 * 30 * 100)
= 35.4564 -->

<!-- ### Web Sockets
(0.325 / 1000000) * 60 * 24 * 30 * 100
+
(1.30 / 1000000) * (60 / 30) * 60 * 5 * 30 * 100
= 3.744 -->

### SQS
(0.50 / 1000000) * (60 / 30) * 60 * 5 * 30 * 100 - 0.5
= 0.4

### HTTP API
(1.29 / 1000000) * (60 / 30) * 60 * 5 * 30 * 100
= 2.322

## Cognito
0.0055 * 100 - 0.0055 * 50000
= 0

## DynamoDB
(1.4231 / 1000000) * (60 / 30) * 60 * 5 * 30 * 100
+
(0.2846 / 1000000) * (60 / 30) * 60 * 5 * 30 * 100
= 3.07386

## Lambda (ARM 128 mb)
(0.20 / 1000000) * (60 / 30) * 60 * 5 * 30 * 100
+
(0.0000000017 / 1000) * 5 * (60 / 30) * 60 * 5 * 30 * 100
= 0.3600153


# Lambdas
## Settings
* **Input**: Media type
* **Output**: Max AFK time

## MediaInfo
* **Input**: Media type, media identifier and optionally display name
* **Output**: Media type, media identifier, display name and hours read that day

## Backfill
* **Input**: JSON data/file or last update date time
* **Output**: JSON data or success message

## StatusUpdate
* **Input**: Date time, media identifier and start/stop boolean indicators
* **Output**: Total hours read for that media that day
*Maybe use SQS queues to buffer and have the lambda take batches every 30 seconds? If so then no output* - asynchronous invocation with destinations


# DynamoDB Tables
## Settings
* PK - User
* SK - MediaType
* Leaderboard
* MaxAFK
* MaxBlur
* InactivityBlurAmount
* MenuBlurAmount
* MaxLoadedLines

## Media
* PK - MediaType # User
* SK - Date # Identifier or Identifier ##################### MIGHT NOT WORK CHECK
* LSI 1 - LastUpdate
* DisplayName (only when Identifer is provided by itself)
* LastRead
* Stats

## Leaderboard
* PK - TimePeriod # MediaType
* SK - User
* MediaNames
* LSI 1 - TimeRead
* LSI 2 - CharsRead




# CDK Bootstrap
`cdk bootstrap aws://868004641356/ap-southeast-2 --profile 868004641356_AdministratorAccess --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess --trust 305354055033 --trust 136138178459`

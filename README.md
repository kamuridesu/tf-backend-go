# TF Backend Go

A jump to the sky turns to a Rider Kick

## Intro

A simple HTTP backend to store Terraform/OpenTofu state information using SQLite or Postgres as database.
It also works as an AWS Lambda Serverless service using API Gateway as HTTP frontend and DynamoDB as storage.

## Build


### Linux
```
CGO_ENABLED=1 go build -ldflags='-s -w -extldflags "-static"' -o tf-backend
```

### Lambda

Just run the `build.sh` script to generate the main.zip file. In case you want to run the commands yourself:
```sh
CGO_ENABLED=0 go build -ldflags='-s -w -extldflags "-static"' -o bootstrap  # Build project
zip main.zip bootstrap  # Stores it into a main.zip file
rm bootstrap  # removes the executable file
```

In case you're using Windows, just follow this documentation: https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html#golang-package-windows

## AWS Setup

### Lambda

Create a new Lambda Function, give it the name you want and as for the runtime select `Amazon Linux 2023`. Choose the arch compatible with your build and click Create Function.

Then, in Configurations -> Environment Variables, add these following variables:
```
ACCESS_KEY=
SECRET_ACCESS_KEY=
AUTH_USERS=
```
The access key and secret access key should be created on AWS. Just follow the docs. For the AUTH_USERS env it should be in this format: USER:PASSWORD;USER2:PASSWORD2

Then just upload the .zip file containing the executable.

### API Gateway

Click Create API, then HTTP API. Add Lambda integration and search for the Lambda Function you created. Select it and give it a name. 

Then create a route that starts with / and another that starts with /{proxy+}, the later will be used to give a name to the Terraform State. The method for all routes should be `ANY`.

After creating the routes just integrate it with the Lambda Function you created. Get the URL clicking on Deploy -> Stages.

### DynamoDB

Create a new table called `tfstates` with Name as Partition Key and Locked as Sort Key. Leave the other options as default.

## Routes

- `/tfstates` -> base route for Terraform Operations
- `tfstates/:name` -> route for Terraform Operations

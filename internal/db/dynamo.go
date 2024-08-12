package db

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func OpenDynamoDB() {
	svc := dynamodb.New(nil)
	svc.
}

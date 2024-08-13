package db

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoDB struct {
	svc *dynamodb.DynamoDB
}

const TABLE_NAME = "tfstates"

func OpenDynamoDB() *DynamoDB {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGIONS")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("ACCESS_KEY"),
			os.Getenv("SECRET_ACCESS_KEY"),
			""),
	}),
	)
	d := DynamoDB{svc: dynamodb.New(sess)}
	d.CreateTableIfNotExists()
	return &d
}

func (d *DynamoDB) ListTables() []string {
	tableNames := make([]string, 0)
	input := &dynamodb.ListTablesInput{}

	for {
		result, err := d.svc.ListTables(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case dynamodb.ErrCodeInternalServerError:
					log.Fatal(fmt.Sprint(dynamodb.ErrCodeInternalServerError, aerr.Error()))
				default:
					log.Fatal(aerr.Error())
				}
			} else {
				log.Fatal(err.Error())
			}
			return tableNames
		}

		for _, n := range result.TableNames {
			tableNames = append(tableNames, *n)
		}

		input.ExclusiveStartTableName = result.LastEvaluatedTableName

		if result.LastEvaluatedTableName == nil {
			break
		}
	}

	return tableNames
}

func (d *DynamoDB) CreateTable() {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Name"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("Locked"),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Name"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("Locked"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
		TableName: aws.String(TABLE_NAME),
	}

	_, err := d.svc.CreateTable(input)
	if err != nil {
		log.Fatalf("Got error calling CreatingTable: %s", err)
	}

	slog.Info(fmt.Sprint("Created the table ", TABLE_NAME))
}

func (d *DynamoDB) CreateTableIfNotExists() {
	tables := d.ListTables()
	for _, table := range tables {
		if table == TABLE_NAME {
			return
		}
	}
	d.CreateTable()
}

func (d *DynamoDB) NewState(state *State) error {
	av, err := dynamodbattribute.MarshalMap(state.AsDTO())
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(TABLE_NAME),
	}

	_, err = d.svc.PutItem(input)
	return err
}

func (d *DynamoDB) GetState(name string) (*State, error) {
	result, err := d.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(name),
			},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling GetItem: %s", err)
	}

	if result.Item == nil {
		return nil, errors.New("Could not find state '" + name + "'")
	}

	statedto := StateDTO{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &statedto)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
	}

	state := State{
		Name:     statedto.Name,
		Contents: statedto.Contents,
		Locked:   statedto.Locked == 1,
	}

	return &state, nil
}

func (d *DynamoDB) UpdateState(state *State) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":r": {
				S: aws.String(state.Contents),
			},
		},
		TableName: aws.String(TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"Name": {
				S: aws.String(state.Name),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set Content = :r"),
	}

	_, err := d.svc.UpdateItem(input)
	return err
}

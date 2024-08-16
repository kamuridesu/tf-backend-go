package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDB struct {
	svc *dynamodb.Client
}

const TABLE_NAME = "tfstates"

func OpenDynamoDB() (*DynamoDB, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("ACCESS_KEY"),
			os.Getenv("SECRET_ACCESS_KEY"),
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	d := DynamoDB{svc: dynamodb.NewFromConfig(cfg)}
	err = d.CreateTableIfNotExists()
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (d *DynamoDB) ListTables() []string {
	tableNames := make([]string, 0)

	var output *dynamodb.ListTablesOutput
	var err error

	tablePaginator := dynamodb.NewListTablesPaginator(d.svc, &dynamodb.ListTablesInput{})
	for tablePaginator.HasMorePages() {
		output, err = tablePaginator.NextPage(context.TODO())
		if err != nil {
			slog.Error(fmt.Sprintf("could not list tables, reason: %v", err))
			break
		} else {
			tableNames = append(tableNames, output.TableNames...)
		}
	}

	return tableNames
}

func (d *DynamoDB) CreateTable() error {

	_, err := d.svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("Name"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("Contents"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("Locked"),
				AttributeType: types.ScalarAttributeTypeN,
			},
		},
		TableName: aws.String(TABLE_NAME),
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})

	if err != nil {
		return err
	} else {
		waiter := dynamodb.NewTableExistsWaiter(d.svc)
		err := waiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
			TableName: aws.String(TABLE_NAME)}, 5*time.Minute)
		if err != nil {
			return err
		}
	}

	return nil

}

func (d *DynamoDB) CreateTableIfNotExists() error {
	tables := d.ListTables()
	for _, table := range tables {
		if table == TABLE_NAME {
			return nil
		}
	}
	return d.CreateTable()
}

func (d *DynamoDB) NewState(state *State) error {

	item, err := attributevalue.MarshalMap(state.AsDTO())
	if err != nil {
		return err
	}
	_, err = d.svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(TABLE_NAME), Item: item,
	})

	return err
}

func (d *DynamoDB) GetState(name string) (*State, error) {
	var err error
	keyEx := expression.Key("Name").Equal(expression.Value(name))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, err
	}

	var states []StateDTO

	queryPaginator := dynamodb.NewQueryPaginator(d.svc, &dynamodb.QueryInput{
		TableName:                 aws.String(TABLE_NAME),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})
	for queryPaginator.HasMorePages() {
		response, err := queryPaginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		var statedto []StateDTO
		err = attributevalue.UnmarshalListOfMaps(response.Items, &statedto)
		if err != nil {
			return nil, err
		}
		states = append(states, statedto...)
	}

	if len(states) == 0 {
		slog.Warn("states is empty")
		return nil, nil
	}

	if len(states) > 1 {
		slog.Warn("more than one state found")
		return nil, fmt.Errorf("length of states inconsistent, try to create a new index")
	}

	statedto := states[0]

	state := State{
		Name:     statedto.Name,
		Contents: statedto.Contents,
		Locked:   statedto.Locked == 1,
	}

	return &state, nil
}

func (d *DynamoDB) UpdateState(state *State) error {
	update := expression.Set(expression.Name("Contents"), expression.Value(state.Contents))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return err
	}

	locked := "0"
	if state.Locked {
		locked = "1"
	}

	_, err = d.svc.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"Name":   &types.AttributeValueMemberS{Value: state.Name},
			"Locked": &types.AttributeValueMemberN{Value: locked},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})

	return err
}

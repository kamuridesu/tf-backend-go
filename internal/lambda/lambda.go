package lambda

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kamuridesu/tf-backend-go/internal/db"
)

var Database *db.Database

var NotFoundResponse = events.APIGatewayProxyResponse{
	StatusCode: 404,
	Body:       "route not found",
}

func BuildApiResponse(status int, msg string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       msg,
	}
}

func BuildResponseFromNillable(status int, err error) events.APIGatewayProxyResponse {
	return BuildApiResponse(status, err.Error())
}

func Router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	targetPath, ok := req.PathParameters["proxy"]

	if ok {
		if !strings.HasPrefix(targetPath, "tfstate") {
			return NotFoundResponse, nil
		}
		parsedPath := strings.Split(targetPath, "/")
		if len(parsedPath) > 2 {
			return NotFoundResponse, nil
		}
		name := parsedPath[1]
		switch req.HTTPMethod {
		case "POST":
			status, err := HandlePost(name, req.Body, Database)
			return BuildResponseFromNillable(status, err), nil
		case "GET":
			status, content, err := HandleGet(name, Database)
			if err != nil {
				return BuildResponseFromNillable(status, err), nil
			}
			return BuildApiResponse(status, content), nil
		case "LOCK":
			status, err := HandleLock(name, Database)
			return BuildResponseFromNillable(status, err), nil
		case "UNLOCK":
			status, err := HandleUnlock(name, Database)
			return BuildResponseFromNillable(status, err), nil
		}

	}
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "ok"}, nil
}

func Main() {
	var err error

	Database, err = db.StartDB(db.DatabaseType("dynamodb"), "")
	if err != nil {
		panic(err)
	}
	lambda.Start(Router)
}

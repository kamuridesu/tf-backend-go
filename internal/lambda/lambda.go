package lambda

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kamuridesu/tf-backend-go/internal/db"
)

var Database *db.Database

var NotFoundResponse = events.APIGatewayProxyResponse{
	StatusCode: 404,
	Body:       "route or state not found",
}

type DefaultResponseWhenNotFound struct {
	Version int `json:"version"`
}

func BuildApiResponse(status int, msg string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       msg,
	}
}

func Router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ev, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	slog.Info(string(ev))
	targetPath, ok := req.PathParameters["proxy"]
	reply := NotFoundResponse
	slog.Info("Received " + targetPath + " as path")
	if ok {
		if !strings.HasPrefix(targetPath, "tfstate") {
			return NotFoundResponse, nil
		}
		parsedPath := strings.Split(targetPath, "/")
		if len(parsedPath) > 2 {
			return NotFoundResponse, nil
		}
		name := parsedPath[1]
		slog.Info("State name: " + name + " with HTTP Method " + req.RequestContext.HTTPMethod)
		switch req.HTTPMethod {
		case "POST":
			status, err := HandlePost(name, req.Body, Database)
			reply = BuildApiResponse(status, err.Error())
		case "GET", "HTTP":
			status, content, err := HandleGet(name, Database)
			if err != nil {
				returnData, err := json.Marshal(&DefaultResponseWhenNotFound{
					Version: 0,
				})
				if err != nil {
					reply = BuildApiResponse(500, err.Error())
				} else {
					reply = BuildApiResponse(status, string(returnData))
				}
			} else {
				reply = BuildApiResponse(status, content)
			}
		case "LOCK":
			status, err := HandleLock(name, Database)
			reply = BuildApiResponse(status, err.Error())
		case "UNLOCK":
			status, err := HandleUnlock(name, Database)
			reply = BuildApiResponse(status, err.Error())
		}

	}
	slog.Info("Reply is " + reply.Body)
	return reply, nil
}

func Main() {
	var err error

	Database, err = db.StartDB(db.DatabaseType("dynamodb"), "")
	if err != nil {
		panic(err)
	}
	lambda.Start(Router)
}

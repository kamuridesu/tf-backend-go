package lambda

import (
	"context"
	"encoding/json"
	"fmt"
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

type DefaultResponseWhenNotFoundS struct {
	Version int `json:"version"`
}

func BuildApiResponse(status int, msg string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       msg,
	}
}

func Router(req events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	reply := NotFoundResponse
	defaultResponseWhenNotFound, err := json.Marshal(&DefaultResponseWhenNotFoundS{
		Version: 0,
	})

	if err != nil {
		reply = BuildApiResponse(500, err.Error())
	}

	targetPath, ok := req.PathParameters["proxy"]
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
		slog.Info("State name: " + name + " with HTTP Method " + req.RequestContext.HTTP.Method)
		switch req.RequestContext.HTTP.Method {
		case "POST":
			status, err := HandlePost(name, req.Body, Database)
			if err != nil {
				reply = BuildApiResponse(status, err.Error())
			}
			reply = BuildApiResponse(status, "ok")
		case "GET":
			status, content, err := HandleGet(name, Database)
			if err != nil {
				reply = BuildApiResponse(status, string(defaultResponseWhenNotFound))
			} else {
				reply = BuildApiResponse(status, content)
			}
		case "LOCK":
			status, err := HandleLock(name, Database)
			if err != nil {
				reply = BuildApiResponse(status, err.Error())
			}
			reply = BuildApiResponse(status, string(defaultResponseWhenNotFound))
		case "UNLOCK":
			status, err := HandleUnlock(name, Database)
			if err != nil {
				reply = BuildApiResponse(status, string(defaultResponseWhenNotFound))
			}
		}

	}
	slog.Info("Reply is " + reply.Body)
	return reply, nil
}

func Logger(c context.Context, m interface{}) {
	fmt.Println(m)
	fmt.Println(c)
}

func Main() {
	var err error

	Database, err = db.StartDB(db.DatabaseType("dynamodb"), "")
	if err != nil {
		panic(err)
	}
	lambda.Start(Router)
}

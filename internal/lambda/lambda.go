package lambda

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/kamuridesu/tf-backend-go/cmd"
	"github.com/kamuridesu/tf-backend-go/internal/db"
)

var (
	Database db.Database

	NotFoundResponse = events.APIGatewayProxyResponse{
		StatusCode: 404,
		Body:       "route or state not found",
	}
	Users, _ = cmd.LoadEnvVars()
)

const (
	PATH_PREFIX string = "tfstates"
)

func ValidateUser(users *[]cmd.User, authData string) bool {
	if authData == "" {
		return false
	}

	if !strings.Contains(authData, "Basic") {
		return false
	}

	decodedData, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authData, "Basic "))
	if err != nil {
		return false
	}

	userData := strings.Split(string(decodedData), ":")
	if len(userData) != 2 {
		return false
	}
	username := userData[0]
	password := userData[1]

	for _, user := range *users {
		if user.Name == username && user.Password == password {
			return true
		}
	}
	return false
}

func BuildApiResponse(status int, msg string, asJson bool) events.APIGatewayProxyResponse {
	body := msg
	if asJson {

		bbody, err := json.Marshal(map[string]string{
			"status": msg,
		})
		if err != nil {
			slog.Error("could not parse msg to json")
			panic(err)
		}
		body = string(bbody)
	}

	headers := map[string]string{
		"content-type": "application/json",
	}

	if !asJson {
		headers["content-type"] = "text/plain; charset=utf-8"
	}
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       body,
		Headers:    headers,
	}
}

func Router(req events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {

	authData := req.Headers["authorization"]

	if !ValidateUser(Users, authData) {
		return BuildApiResponse(http.StatusUnauthorized, "User not authorized", false), nil
	}

	reply := NotFoundResponse

	targetPath := strings.TrimPrefix(req.RawPath, "/")
	targetPath = strings.TrimPrefix(targetPath, PATH_PREFIX)
	parsedPath := strings.Split(targetPath, "/")
	if len(parsedPath) > 2 {
		slog.Error("invalid path")
		return NotFoundResponse, nil
	}
	name := parsedPath[1]
	slog.Info("State name: " + name + " with HTTP Method " + req.RequestContext.HTTP.Method)
	switch req.RequestContext.HTTP.Method {
	case "POST":
		status, err := HandlePost(name, req.Body, Database)
		if err != nil {
			reply = BuildApiResponse(status, err.Error(), true)
		} else {
			reply = BuildApiResponse(status, "ok", true)
		}
	case "GET":
		status, content, err := HandleGet(name, Database)
		if err != nil {
			reply = BuildApiResponse(status, err.Error(), true)
		} else {
			reply = BuildApiResponse(status, content, false)
		}
	case "LOCK":
		status, err := HandleLock(name, Database)
		if err != nil {
			reply = BuildApiResponse(status, err.Error(), true)
		} else {
			reply = BuildApiResponse(status, "ok", true)
		}
	case "UNLOCK":
		status, err := HandleUnlock(name, Database)
		if err != nil {
			reply = BuildApiResponse(status, "ok", true)
		}
	}
	slog.Info(fmt.Sprintf("Returning %s with status %d", reply.Body, reply.StatusCode))
	return reply, nil
}

func Main() {
	var err error

	Database, err = db.NewDatabase(db.Dynamo, "")
	if err != nil {
		panic(err)
	}
	lambda.Start(Router)
}

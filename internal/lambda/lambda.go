package lambda

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// type Event struct {
// 	Name string `json:"name"`
// }

// func HandleRequest(ctx context.Context, event *Event) (*string, error) {
// 	if event == nil {
// 		return nil, fmt.Errorf("rcvd nil event")
// 	}
// 	message := fmt.Sprintf("Hello %s!", event.Name)
// 	return &message, nil
// }

func Router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// if !strings.Contains(req.Path, "tfstates") {
	// 	return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Missing state name"}, fmt.Errorf("Missing state name")
	// }

	// stateName = ""
	log.Println(req.Path)
	log.Println(req.PathParameters)
	log.Println(req.Body)
	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "ok"}, nil
}

func Main() {
	lambda.Start(Router)
}

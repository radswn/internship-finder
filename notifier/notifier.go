package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type NotificationEvent struct {
	Input string `json:"input"`
}

type Notification struct {
	Message string `json:"message"`
}

func HandleNotifier(event NotificationEvent) (Notification, error) {
	return Notification{Message: "Hello with the " + event.Input}, nil
}

func main() {
	lambda.Start(HandleNotifier)
}

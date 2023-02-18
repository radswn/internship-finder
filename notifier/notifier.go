package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/url"
	"os"
)

type SqsApi struct {
	Client   *sqs.Client
	QueueUrl string
}

type Response struct {
	Offers []Offer `json:"offers"`
}

type Offer struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	Location string `json:"location"`
}

var sqsApi SqsApi

func getSqsApi() SqsApi {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sqs.NewFromConfig(cfg)
	queue := os.Getenv("QueueName")

	result, err := client.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: &queue,
	})

	if err != nil {
		panic("Queue URL retrieval error, " + err.Error())
	}
	queueUrl := *result.QueueUrl

	return SqsApi{
		Client:   client,
		QueueUrl: queueUrl,
	}
}

func HandleNotifier() (Response, error) {
	offers := make([]Offer, 0)
	telegramApi := "https://api.telegram.org/bot" + os.Getenv("BotToken") + "/sendMessage"

	for {
		messageOutput, err := sqsApi.Client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &sqsApi.QueueUrl,
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     3,
		})
		if err != nil {
			panic("error while retrieving message, " + err.Error())
		}

		if len(messageOutput.Messages) == 0 {
			break
		}

		message := messageOutput.Messages[0]
		body := *message.Body
		receiptHandle := message.ReceiptHandle

		var offer Offer
		json.Unmarshal([]byte(body), &offer)

		offers = append(offers, offer)

		_, err = sqsApi.Client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
			QueueUrl:      &sqsApi.QueueUrl,
			ReceiptHandle: receiptHandle,
		})

		if err != nil {
			panic("error while deleting message, " + err.Error())
		}

		_, err = http.PostForm(
			telegramApi,
			url.Values{
				"chat_id": {os.Getenv("ChatID")},
				"text":    {fmt.Sprintf("%s\n%s\n\n%s", offer.Title, offer.Location, offer.Link)},
			})

		if err != nil {
			log.Printf("error when posting text to the chat: %s", err.Error())
			return Response{}, err
		}
	}

	return Response{Offers: offers}, nil
}

func main() {
	if len(os.Getenv("BotToken"))*len(os.Getenv("ChatID")) == 0 {
		godotenv.Load()
	}

	sqsApi = getSqsApi()
	lambda.Start(HandleNotifier)
}

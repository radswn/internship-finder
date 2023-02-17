package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gocolly/colly"
	"os"
	"strconv"
	"strings"
	"time"
)

type SqsApi struct {
	Client   *sqs.Client
	QueueUrl string
}

type Event struct {
	Site string `json:"site"`
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

func isOfferFromToday(offerDate string) bool {
	current := time.Now()

	currentDay := strconv.Itoa(current.Day())
	currentMonthShort := current.Month().String()[:3]

	return strings.Contains(offerDate, currentMonthShort+" "+currentDay)
}

func HandleChecker(event Event) (Response, error) {
	c := colly.NewCollector(colly.AllowedDomains("jobs.apple.com"))

	offers := make([]Offer, 0)

	c.OnHTML("td.table-col-1", func(e *colly.HTMLElement) {
		if !isOfferFromToday(e.ChildText("span.table--advanced-search__date")) {
			return
		}

		title := e.ChildText("a")
		link := constructLink(event.Site, e.ChildAttr("a", "href"))
		location := e.DOM.SiblingsFiltered("td.table-col-2").Text()

		offer := Offer{
			Title:    title,
			Link:     link,
			Location: location,
		}

		offers = append(offers, offer)
	})

	err := c.Visit(event.Site)

	if err == nil {
		smInput := &sqs.SendMessageInput{
			MessageBody: aws.String("Hello there"),
			QueueUrl:    &sqsApi.QueueUrl,
		}
		_, err_ := sqsApi.Client.SendMessage(context.TODO(), smInput)
		if err_ != nil {
			panic("error while sending message, " + err_.Error())
		}
	}

	return Response{Offers: offers}, err
}

func constructLink(visitedUrl string, href string) (url string) {
	baseUrlEndIdx := strings.Index(visitedUrl, ".com") + 4
	baseUrl := visitedUrl[:baseUrlEndIdx]
	return baseUrl + href
}

func main() {
	sqsApi = getSqsApi()
	lambda.Start(HandleChecker)
}

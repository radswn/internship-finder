package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
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
	Date string `json:"date"`
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

func isOfferDateRelevant(offerDate string, requestedDate string) bool {
	if requestedDate == "" {
		current := time.Now()

		currentDay := strconv.Itoa(current.Day())
		currentMonthShort := current.Month().String()[:3]

		return strings.Contains(offerDate, currentMonthShort+" "+currentDay)
	}

	return strings.Contains(offerDate, requestedDate)
}

func HandleChecker(event Event) (Response, error) {
	c := colly.NewCollector(colly.AllowedDomains("jobs.apple.com"))

	offers := make([]Offer, 0)

	c.OnHTML("td.table-col-1", func(e *colly.HTMLElement) {
		if !isOfferDateRelevant(e.ChildText("span.table--advanced-search__date"), event.Date) {
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

	if err == nil && sqsApi.Client != nil {
		for _, o := range offers {
			data, _ := json.Marshal(o)
			s := string(data)

			smInput := &sqs.SendMessageInput{
				MessageBody: &s,
				QueueUrl:    &sqsApi.QueueUrl,
			}

			_, err_ := sqsApi.Client.SendMessage(context.TODO(), smInput)
			if err_ != nil {
				panic("error while sending message, " + err_.Error())
			}
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

package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gocolly/colly"
	"io"
	"net/http"
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
	AmazonSite string `json:"amazonSite"`
	AppleSite  string `json:"appleSite"`
	Date       string `json:"date"`
}

type AmazonJobs struct {
	Jobs []AmazonJob `json:"jobs"`
}

type AmazonJob struct {
	Location   string `json:"normalized_location"`
	JobPath    string `json:"job_path"`
	Title      string `json:"title"`
	PostedDate string `json:"posted_date"`
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

func isOfferDateRelevant(offerDate string, requestedDate string, dateFormat string) bool {
	if requestedDate == "all" {
		return true
	}

	var current time.Time

	if requestedDate == "" {
		current = time.Now()
	} else {
		var err error
		current, err = time.Parse("02.01", requestedDate)
		if err != nil {
			panic("error while parsing date" + err.Error())
		}
	}

	currentDay := strconv.Itoa(current.Day())
	currentMonth := current.Month().String()

	switch dateFormat {
	case "apple":
		requestedDate = currentMonth[:3] + " " + currentDay
	case "amazon":
		requestedDate = currentMonth + " " + currentDay
	}

	return strings.Contains(offerDate, requestedDate)
}

func HandleChecker(event Event) (Response, error) {
	offers, err := collectAllOffers(event)

	if err == nil && sqsApi.Client != nil {
		forwardOffersToQueue(&offers)
	}

	return Response{Offers: offers}, err
}

func collectAllOffers(event Event) ([]Offer, error) {
	offers := make([]Offer, 0)
	var err error

	if event.AppleSite != "" {
		appleOffers, err := getAppleOffers(event)
		if err == nil {
			offers = append(offers, *appleOffers...)
		}
	}

	if event.AmazonSite != "" {
		amazonOffers, err := getAmazonOffers(event)
		if err == nil {
			offers = append(offers, *amazonOffers...)
		}
	}

	return offers, err
}

func getAppleOffers(event Event) (*[]Offer, error) {
	c := colly.NewCollector()
	offers := make([]Offer, 0)

	c.OnHTML("td.table-col-1", func(e *colly.HTMLElement) {
		if !isOfferDateRelevant(e.ChildText("span.table--advanced-search__date"), event.Date, "apple") {
			return
		}

		title := e.ChildText("a")
		link := constructLink(event.AppleSite, e.ChildAttr("a", "href"))
		location := e.DOM.SiblingsFiltered("td.table-col-2").Text()

		offer := Offer{
			Title:    title,
			Link:     link,
			Location: location,
		}

		offers = append(offers, offer)
	})
	err := c.Visit(event.AppleSite)

	return &offers, err
}

func getAmazonOffers(event Event) (*[]Offer, error) {
	offers := make([]Offer, 0)

	res, err := http.Get(event.AmazonSite)
	if err != nil {
		panic("error while GETting Amazon website" + err.Error())
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic("error reading response body" + err.Error())
	}

	var data AmazonJobs

	err = json.Unmarshal(resBody, &data)
	if err != nil {
		panic("error unmarshalling json" + err.Error())
	}

	for _, j := range data.Jobs {
		if !isOfferDateRelevant(j.PostedDate, event.Date, "amazon") {
			continue
		}

		offer := Offer{
			Title:    j.Title,
			Link:     constructLink(event.AmazonSite, j.JobPath),
			Location: j.Location,
		}

		offers = append(offers, offer)
	}

	return &offers, err
}

func constructLink(visitedUrl string, href string) (url string) {
	baseUrlEndIdx := strings.Index(visitedUrl, ".jobs") + 5

	if baseUrlEndIdx == 4 {
		baseUrlEndIdx = strings.Index(visitedUrl, ".com") + 4
	}
	baseUrl := visitedUrl[:baseUrlEndIdx]
	return baseUrl + href
}

func forwardOffersToQueue(offers *[]Offer) {
	for _, o := range *offers {
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

func main() {
	sqsApi = getSqsApi()
	lambda.Start(HandleChecker)
}

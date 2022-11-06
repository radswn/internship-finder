package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
)

type Event struct {
	Site string `json:"site"`
}

type Response struct {
	Offers []Offer `json:"offers"`
}

type Offer struct {
	Title string `json:"title"`
}

func HandleLambdaEvent(event Event) (Response, error) {
	c := colly.NewCollector(
		colly.AllowedDomains("jobs.apple.com"),
		colly.AllowURLRevisit())

	offers := make([]Offer, 0)

	c.OnHTML("td.table-col-1", func(e *colly.HTMLElement) {
		offer := Offer{Title: e.ChildText("a")}
		offers = append(offers, offer)
	})

	err := c.Visit(event.Site)
	return Response{Offers: offers}, err
}

func main() {
	lambda.Start(HandleLambdaEvent)
}

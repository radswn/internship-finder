package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
)

type MyEvent struct {
	Site string `json:"site"`
}

type MyResponse struct {
	Offers []Offer `json:"offers"`
}

type Offer struct {
	Title string `json:"title"`
}

func HandleLambdaEvent(event MyEvent) (MyResponse, error) {
	c := colly.NewCollector(
		colly.AllowedDomains("jobs.apple.com"),
		colly.AllowURLRevisit())

	offers := make([]Offer, 0)

	c.OnHTML("td.table-col-1", func(e *colly.HTMLElement) {
		offer := Offer{Title: e.ChildText("a")}
		offers = append(offers, offer)
	})

	c.Visit("https://jobs.apple.com/en-us/search?team=internships-STDNT-INTRN")
	return MyResponse{Offers: offers}, nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}

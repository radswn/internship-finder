package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
	"strings"
)

type Event struct {
	Site string `json:"site"`
}

type Response struct {
	Offers []Offer `json:"offers"`
}

type Offer struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func HandleLambdaEvent(event Event) (Response, error) {
	c := colly.NewCollector(
		colly.AllowedDomains("jobs.apple.com"),
		colly.AllowURLRevisit())

	offers := make([]Offer, 0)

	c.OnHTML("td.table-col-1", func(e *colly.HTMLElement) {
		title := e.ChildText("a")
		link := constructLink(event.Site, e.ChildAttr("a", "href"))

		offer := Offer{
			Title: title,
			Link:  link,
		}
		offers = append(offers, offer)
	})

	err := c.Visit(event.Site)
	return Response{Offers: offers}, err
}

func constructLink(visitedUrl string, href string) (url string) {
	baseUrlEndIdx := strings.Index(visitedUrl, ".com") + 4
	baseUrl := visitedUrl[:baseUrlEndIdx]
	return baseUrl + href
}

func main() {
	lambda.Start(HandleLambdaEvent)
}

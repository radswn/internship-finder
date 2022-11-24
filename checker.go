package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
	"strconv"
	"strings"
	"time"
)

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

func isOfferFromToday(offerDate string) bool {
	current := time.Now()

	currentDay := strconv.Itoa(current.Day())
	currentMonthShort := current.Month().String()[:3]

	return strings.Contains(offerDate, currentMonthShort+" "+currentDay)
}

func HandleLambdaEvent(event Event) (Response, error) {
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

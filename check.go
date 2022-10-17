package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
)

type MyEvent struct {
	Site string `json:"site"`
}

type MyResponse struct {
	Message string `json:"message"`
}

func HandleLambdaEvent(event MyEvent) (MyResponse, error) {
	c := colly.NewCollector(
		colly.AllowedDomains("hackerspaces.org", "wiki.hackerspaces.org"),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://hackerspaces.org
	c.Visit("https://hackerspaces.org/")
	return MyResponse{Message: fmt.Sprintf("You wanted to scrape %s, so there you go...", event.Site)}, nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}

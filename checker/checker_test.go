package main

import (
	"net/url"
	"testing"
)

var site = "https://jobs.apple.com/en-us/search?team=internships-STDNT-INTRN"
var offers []Offer

func TestConnection(t *testing.T) {
	event := Event{site}
	resp, err := HandleChecker(event)

	if err != nil {
		t.Fatalf(`Received error %q`, err)
	}

	offers = resp.Offers
}

func TestNonEmptyTitles(t *testing.T) {
	for i, o := range offers {
		if o.Title == "" {
			t.Fatalf(`Offer number %d got an empty title`, i)
		}
	}
}

func TestNonEmptyLocation(t *testing.T) {
	for i, o := range offers {
		if o.Location == "" {
			t.Fatalf(`Offer number %d got an empty location`, i)
		}
	}
}

func TestValidLinks(t *testing.T) {
	for i, o := range offers {
		u, err := url.ParseRequestURI(o.Link)
		if err != nil {
			t.Fatalf(`Invalid url %q for offer number %d`, u, i)
		}
	}
}

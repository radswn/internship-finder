package main

import (
	"testing"
)

var site = "https://jobs.apple.com/en-us/search?team=internships-STDNT-INTRN"

func TestNonEmptyArray(t *testing.T) {
	event := Event{site}
	resp, err := HandleLambdaEvent(event)

	if err != nil {
		t.Fatalf(`Received error %q`, err)
	}

	if len(resp.Offers) == 0 {
		t.Fatalf(`Expected a non-empty array, received an empty one`)
	}
}

func TestNonEmptyTitles(t *testing.T) {
	event := Event{site}
	resp, err := HandleLambdaEvent(event)

	if err != nil {
		t.Fatalf(`Received error %q`, err)
	}

	for i, o := range resp.Offers {
		if o.Title == "" {
			t.Fatalf(`Offer number %d got an empty title`, i)
		}
	}
}

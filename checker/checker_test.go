package main

import (
	"net/url"
	"testing"
)

var appleSite = "https://jobs.apple.com/en-us/search?team=internships-STDNT-INTRN"
var amazonSite = "https://www.amazon.jobs/en/search.json?category%5B%5D=software-development&category%5B%5D=machine-learning-science&is_intern%5B%5D=1&radius=24km&facets%5B%5D=normalized_country_code&facets%5B%5D=normalized_state_name&facets%5B%5D=normalized_city_name&facets%5B%5D=location&facets%5B%5D=business_category&facets%5B%5D=category&facets%5B%5D=schedule_type_id&facets%5B%5D=employee_class&facets%5B%5D=normalized_location&facets%5B%5D=job_function_id&facets%5B%5D=is_manager&facets%5B%5D=is_intern&offset=0&result_limit=100&sort=recent&latitude=&longitude=&loc_group_id=&loc_query=&base_query=&city=&country=&region=&county=&query_options=&business_category%5B%5D=student-programs&category%5B%5D=software-development&category%5B%5D=machine-learning-science&"
var date = ""
var offers []Offer

func TestConnection(t *testing.T) {
	event := Event{amazonSite, appleSite, date}
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

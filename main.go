package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

type ResponseHeader struct {
	Response Response `json:"response"`
}

type Response struct {
	NumFound int64      `json:"numFound"`
	Docs     []ShoeItem `json:"docs"`
}

type ShoeItem struct {
	ParentName string  `json:"parent_name"`
	Pcode      string  `json:"pcode"`
	PriceHigh  float64 `json:"price_high"`
	PriceLow   float64 `json:"price_low"`
}

type QueryResult struct {
	QueryURL                string
	ThresholdPrice          float64
	MatchingShoes           []ShoeItem
	ShoesAtOrBelowThreshold []ShoeItem
}

type Email struct {
	body         string
	fromAddress  string
	fromPassword string
	toAddress    string
}

// getShoesAtOrBelowThreshold() takes a []ShoeItem and a threshold value and
// returns a sorted (low to high) []ShoeItem of shoes at or below the threshold.
func getShoesAtOrBelowThreshold(shoes []ShoeItem, threshold float64) []ShoeItem {
	meetThreshold := []ShoeItem{}

	// Filter out shoes that exceed the threshold.
	for _, shoe := range shoes {
		if shoe.PriceLow <= threshold {
			meetThreshold = append(meetThreshold, shoe)
		}
	}

	// Sort threshold items by lowest to greatest price.
	sort.Slice(meetThreshold, func(i, j int) bool {
		return meetThreshold[i].PriceLow < meetThreshold[j].PriceLow
	})

	return meetThreshold
}

// createEmailBody() formats queries with matching results for emailing using
// a Go template.
func createEmailBody(queryResults []QueryResult) string {
	htmlBody := new(bytes.Buffer)
	tmpl, err := template.ParseFiles("email.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(htmlBody, queryResults)

	return htmlBody.String()
}

// sendEmail() sends sends a single email with results for queries that met
// the criteria.
func sendEmail(email Email) {
	emailAuth := smtp.PlainAuth("", email.fromAddress, email.fromPassword, "smtp.gmail.com")
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: SHOE ALERT\n"

	msg := []byte(subject + mime + "\n" + email.body)

	err := smtp.SendMail("smtp.gmail.com:587", emailAuth, email.fromAddress, []string{email.toAddress}, msg)
	if err != nil {
		log.Fatalf("smtp error: %s", err)
	}
}

// getQueryURLs() takes a comma separated string of URLs and return them
// as a []string.
func getQueryURLs(queryURLs string) []string {
	urls := strings.Split(queryURLs, ",")

	for idx, url := range urls {
		urls[idx] = strings.Trim(url, " ")
	}

	return urls
}

// getQueryMatches() gets all mathing ShoeItems from a queryURL.
func getQueryMatches(queryURL string) []ShoeItem {
	// Get data from the API.
	req, err := http.NewRequest(http.MethodGet, queryURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// Read the respponse.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the JSON for processing.
	var responseHeader ResponseHeader
	if err := json.Unmarshal(body, &responseHeader); err != nil {
		log.Fatal(err)
	}

	// Carve out all the matching ShoeItem and return them.
	return responseHeader.Response.Docs
}

func main() {
	// Load settings
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	queryURLString := os.Getenv("QUERY_URL")
	toAddress := os.Getenv("RECIPIENT_EMAIL")
	fromAddress := os.Getenv("FROM_GMAIL")
	fromPassword := os.Getenv("FROM_GMAIL_APP_PASSWORD")
	tp := os.Getenv("THRESHOLD_PRICE")
	thresholdPrice, err := strconv.ParseFloat(tp, 64)
	if err != nil {
		log.Fatal(err)
	}

	// Gather email information in one place.
	email := Email{
		fromAddress:  fromAddress,
		fromPassword: fromPassword,
		toAddress:    toAddress,
	}

	// Get the individual query URLs.
	queryURLs := getQueryURLs(queryURLString)

	// Process each query.
	queryResults := []QueryResult{}
	for _, queryURL := range queryURLs {
		matchingShoes := getQueryMatches(queryURL)

		queryResult := QueryResult{
			QueryURL:       queryURL,
			ThresholdPrice: thresholdPrice,
			MatchingShoes:  matchingShoes,
		}

		// Filter out non-matching shoes.
		queryResult.ShoesAtOrBelowThreshold = getShoesAtOrBelowThreshold(queryResult.MatchingShoes, queryResult.ThresholdPrice)

		// No filtered shoes matching the query, so don't add it to the results.
		if len(queryResult.ShoesAtOrBelowThreshold) <= 0 {
			continue
		}

		queryResults = append(queryResults, queryResult)
	}

	// Email out the matching and sorted shoes
	if len(queryResults) > 0 {
		email.body = createEmailBody(queryResults)
		sendEmail(email)
	}
}

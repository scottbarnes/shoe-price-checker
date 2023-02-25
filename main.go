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

type ShoeData struct {
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
// returns a sorted (low to high) []ShoeItem of ShoeItem at or below the threshold.
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

// Format matching shoes in HTML for emailing
func createEmailBody(shoeData ShoeData) string {
	htmlBody := new(bytes.Buffer)
	tmpl, err := template.ParseFiles("email.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(htmlBody, shoeData)

	return htmlBody.String()
}

// Email the filtered and sorted shoes.
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

func main() {
	// Load settings
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	queryURL := os.Getenv("QUERY_URL")
	tp := os.Getenv("THRESHOLD_PRICE")
	thresholdPrice, err := strconv.ParseFloat(tp, 64)
	if err != nil {
		log.Fatal(err)
	}
	toAddress := os.Getenv("RECIPIENT_EMAIL")
	fromAddress := os.Getenv("FROM_GMAIL")
	fromPassword := os.Getenv("FROM_GMAIL_APP_PASSWORD")

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

	// Gather the relevant data in one place for convenience.
	shoeData := ShoeData{
		QueryURL:       queryURL,
		ThresholdPrice: thresholdPrice,
		MatchingShoes:  responseHeader.Response.Docs,
	}

	email := Email{
		fromAddress:  fromAddress,
		fromPassword: fromPassword,
		toAddress:    toAddress,
	}

	// Filter out non-matching shoes.
	shoeData.ShoesAtOrBelowThreshold = getShoesAtOrBelowThreshold(shoeData.MatchingShoes, shoeData.ThresholdPrice)

	// No filtered shoes matching the query, so exit.
	if len(shoeData.ShoesAtOrBelowThreshold) <= 0 {
		os.Exit(0)
	}

	// Email out the matching and sorted shoes
	email.body = createEmailBody(shoeData)
	sendEmail(email)
}

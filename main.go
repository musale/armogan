package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

// armoganURL is the website with the watches that I want
const armoganURL = "https://www.armogan.com/us/all-watches-straps/watches/spirit-of-st-louis"

// currentPrice is the price I don't want to pay
const currentPrice = 225

// africasTalkingEndpoint is the AfricasTalking Endpoint
const africasTalkingEndpoint = "https://api.africastalking.com/version1/messaging"

// ArmoganWatch contains the watch data I need to fetch from the site
type ArmoganWatch struct {
	name, photoURL string
	price          float64
}

// checkPrice will return True if the price has changed from currentPrice
func (k ArmoganWatch) checkPrice() bool {
	return k.price < currentPrice
}

// sms will send me an SMS
func sms(watches []ArmoganWatch) error {
	apiKey := os.Getenv("AT_API_KEY")
	username := os.Getenv("AT_USERNAME")
	to := os.Getenv("MY_PHONENUMBER")
	message := "some text here"

	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"to":       to,
		"message":  message,
	})
	log.Println((string(requestBody)))
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", africasTalkingEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("apiKey", apiKey)

	client := http.Client{}
	resp, err := client.Do(request)

	if err != nil {
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("RESPONSE", string(body))
	return nil
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}
	// Get the armoganURL HTML
	resp, err := http.Get(armoganURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var armogans, changedPrice []ArmoganWatch
	doc.Find(".product-item-info").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		photoURL, _ := s.Find("img.product-image-photo").Attr("src")
		name := s.Find("a.product-item-link").Text()
		price := s.Find("span.price").Text()
		price = strings.Replace(price, "$", "", -1)
		priceAmount, _ := strconv.ParseFloat(price, 64)

		armogans = append(armogans, ArmoganWatch{name: name, photoURL: photoURL, price: priceAmount})
	})

	for _, a := range armogans {
		if !a.checkPrice() {
			changedPrice = append(changedPrice, a)
		}
	}

	// email(changedPrice)
	sms(changedPrice)
}

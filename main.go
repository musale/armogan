package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// armoganURL is the website with the watches that I want
const armoganURL = "https://www.armogan.com/us/all-watches-straps/watches/spirit-of-st-louis"

// currentPrice is the price I don't want to pay
const currentPrice = 158

// africasTalkingEndpoint is the AfricasTalking Endpoint
const africasTalkingEndpoint = "https://api.africastalking.com/version1/messaging"

// ArmoganWatch contains the watch data I need to fetch from the site
type ArmoganWatch struct {
	name, photoURL string
	price          float64
}

// changedPrice will return True if the price has changed from currentPrice
func (k ArmoganWatch) changedPrice() bool {
	return k.price < currentPrice
}

// sms will send me an SMS
func sms(watches []ArmoganWatch) error {
	apiKey := os.Getenv("AT_API_KEY")
	username := os.Getenv("AT_USERNAME")
	to := os.Getenv("MY_PHONENUMBER")
	message := ""

	for _, watch := range watches {
		message += fmt.Sprintf("%s-($%.f)\n", watch.name, watch.price)
	}

	values := url.Values{}
	values.Set("username", username)
	values.Set("to", to)
	values.Set("message", message)

	reader := strings.NewReader(values.Encode())

	req, err := http.NewRequest(http.MethodPost, africasTalkingEndpoint, reader)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Length", strconv.Itoa(reader.Len()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	defer req.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(body))
	return nil
}
func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Println(err)
	// }
	// Get the armoganURL HTML
	log.Println("Getting the price changes on Armogan")
	resp, err := http.Get(armoganURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var armogans, changedPrices []ArmoganWatch
	doc.Find(".product-item-info").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		photoURL, _ := s.Find("img.product-image-photo").Attr("src")
		name := s.Find("a.product-item-link").Text()
		price := s.Find(".special-price").Text()
		if price != "" {
			name = strings.TrimSpace(name)
			price = strings.TrimSpace(strings.Replace(price, "$", "", -1))
			priceAmount, _ := strconv.ParseFloat(price, 64)

			armogans = append(armogans, ArmoganWatch{name: name, photoURL: photoURL, price: priceAmount})
		}
	})

	for _, a := range armogans {
		if a.changedPrice() {
			changedPrices = append(changedPrices, a)
		}
	}
	if len(changedPrices) > 0 {
		sms(changedPrices)
	}
	log.Println("Done getting the price changes on Armogan")
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

// SITEURL is the website with the watches that I want
const SITEURL = "https://www.armogan.com/us/all-watches-straps/watches/spirit-of-st-louis"

// CURRENTPRICE is the price I don't want to pay
const CURRENTPRICE = 225

// AT_ENDPOINT is the AfricasTalking Endpoint
const AT_ENDPOINT = "https://api.africastalking.com/version1/messaging"

// ArmoganWatch contains the watch data I need to fetch from the site
type ArmoganWatch struct {
	name, photoURL string
	price          float64
}

// checkPrice will return True if the price has changed from CURRENTPRICE
func (k ArmoganWatch) checkPrice() bool {
	return k.price < CURRENTPRICE
}

// smtpServer data to smtp server
type smtpServer struct {
	host string
	port string
}

// Address URI to smtp server
func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

// email will email me the watches whose prices have changed
func email(a []ArmoganWatch) error {
	if len(a) > 0 {
		from := os.Getenv("EMAIL_ADDRESS")
		password := os.Getenv("EMAIL_PASSWORD")
		to := []string{from}
		// smtp server configuration.
		smtpServer := smtpServer{host: "smtp.gmail.com", port: "587"}
		// Message.
		mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		subject := "Subject: Test email from Go!\n"
		message := []byte(subject + mime + "<html><body><h1>Hello World!</h1></body></html>")
		// Authentication.
		auth := smtp.PlainAuth("", from, password, smtpServer.host)
		// Sending email.
		err := smtp.SendMail(smtpServer.Address(), auth, from, to, message)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		fmt.Println("Email Sent!")

	}
	return nil
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

	request, err := http.NewRequest("POST", AT_ENDPOINT, bytes.NewBuffer(requestBody))
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
	// Get the SITEURL HTML
	resp, err := http.Get(SITEURL)
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

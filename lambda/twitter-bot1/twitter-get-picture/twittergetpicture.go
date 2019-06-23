package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dbgeek/oauth"
)

type (
	// DirectMessageEvent payload from twitter
	DirectMessageEvent struct {
		MediaID     int64  `json:"mediaID"`
		MediaURL    string `json:"media_url"`
		URL         string `json:"url"`
		MessageText string `json:"message_text"`
		SenderID    string `json:"sender_id"`
		Text        string `json:"text"`
	}
	// Event to send between step functions
	Event struct {
		DirectMessageEvents []DirectMessageEvent `json:"direct-message-events"`
		PictureExists       bool                 `json:"picture-exists"`
	}
)

var (
	consumerSecret, consumerKey, oauthSecret, oauthToken string
	client                                               *http.Client
)

func init() {

	var err error

	consumerKey = os.Getenv("CONSUMER_KEY")
	consumerSecret = os.Getenv("CONSUMER_SECRET_KEY")
	oauthToken = os.Getenv("OAUTH_TOKEN")
	oauthSecret = os.Getenv("OAUTH_SECRET")

	c := oauth.NewConsumer(
		consumerKey,
		consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})

	accessToken := oauth.AccessToken{
		Token:  oauthToken,
		Secret: oauthSecret,
	}

	client, err = c.MakeHttpClient(&accessToken)
	if err != nil {
		log.Fatal(err)
	}
}

// Handler function for the lambda
func Handler(event Event) (Event, error) {

	for _, v := range event.DirectMessageEvents {

		resp, err := client.Get(v.MediaURL)
		if err != nil {
			log.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		fmt.Printf("image file to download is %s and with length: %v \n", v.MediaURL, len(string(body)))
	}

	return event, nil
}

func main() {
	lambda.Start(Handler)
}

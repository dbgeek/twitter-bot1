package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/dbgeek/oauth"
)

type (
	// DirectMessageEvent ..
	DirectMessageEvent struct {
		CreateTimestamp int64  `json:"create_timestamp"`
		MediaID         string `json:"mediaID"`
		MediaURL        string `json:"media_url"`
		URL             string `json:"url"`
		MessageText     string `json:"message_text"`
		SenderID        string `json:"sender_id"`
		Text            string `json:"text"`
		S3bucket        string `json:"s3_bucket"`
		S3path          string `json:"s3_path"`
		Faces           []*rekognition.FaceDetail
	}
	// Event event out from this function to next step
	Event struct {
		DirectMessageEvents []DirectMessageEvent `json:"direct-message-events"`
	}
	// DMReplyEvent Payload
	DMReplyEvent struct {
		Event ReplyEvent `json:"event"`
	}
	// ReplyEvent ..
	ReplyEvent struct {
		Type          string        `json:"type"`
		MessageCreate MessageCreate `json:"message_create"`
	}
	// MessageCreate ..
	MessageCreate struct {
		Target      Target      `json:"target"`
		MessageData MessageData `json:"message_data"`
	}
	// Target ..
	Target struct {
		RecipientID string `json:"recipient_id"`
	}
	// MessageData ..
	MessageData struct {
		Text string `json:"text"`
	}

	face struct {
		emotions string
		gender   string
		ageLow   int64
		ageHigh  int64
	}

	faces []face
)

var (
	consumerSecret string
	consumerKey    string
	oauthSecret    string
	oauthToken     string
	destBucket     string
	client         *http.Client
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

// Handler lambda handler function
func Handler(events Event) error {
	fs := make(faces, 0)
	for _, v := range events.DirectMessageEvents[0].Faces {
		f := face{
			ageLow:  *v.AgeRange.Low,
			ageHigh: *v.AgeRange.High,
			gender:  *v.Gender.Value,
		}
		var c float64
		for _, vv := range v.Emotions {
			if *vv.Confidence > c {
				c = *vv.Confidence
				f.emotions = *vv.Type
			}
		}
		fs = append(fs, f)
	}

	replyMessage := fmt.Sprintf(`Only return first face it detect
face:0, 
age between %d and %d
gender: %s
emotion: %s`, fs[0].ageLow, fs[0].ageHigh, fs[0].gender, fs[0].emotions)

	replyEvent := DMReplyEvent{
		Event: ReplyEvent{
			Type: "message_create",
			MessageCreate: MessageCreate{
				Target: Target{
					RecipientID: events.DirectMessageEvents[0].SenderID,
				},
				MessageData: MessageData{
					Text: string(replyMessage),
				},
			},
		},
	}
	payLoad, _ := json.Marshal(replyEvent)
	fmt.Printf("payload: %v\n", string(payLoad))
	client.Post(
		"https://api.twitter.com/1.1/direct_messages/events/new.json",
		"application/json",
		bytes.NewReader(payLoad),
	)

	return nil
}

func main() {
	lambda.Start(Handler)
}

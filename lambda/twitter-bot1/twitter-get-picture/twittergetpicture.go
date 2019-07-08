package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dbgeek/oauth"
)

type (
	// DirectMessageEvent payload from twitter
	DirectMessageEvent struct {
		CreateTimestamp int64  `json:"create_timestamp"`
		MediaID         string `json:"mediaID"`
		MediaURL        string `json:"media_url"`
		URL             string `json:"url"`
		MessageText     string `json:"message_text"`
		SenderID        string `json:"sender_id"`
		Text            string `json:"text"`
	}
	// Event to send between step functions
	Event struct {
		DirectMessageEvents []DirectMessageEvent `json:"direct-message-events"`
		PictureExists       bool                 `json:"picture-exists"`
	}
	// OutDirectMessageEvent ..
	OutDirectMessageEvent struct {
		CreateTimestamp int64  `json:"create_timestamp"`
		MediaID         string `json:"mediaID"`
		MediaURL        string `json:"media_url"`
		URL             string `json:"url"`
		MessageText     string `json:"message_text"`
		SenderID        string `json:"sender_id"`
		Text            string `json:"text"`
		S3bucket        string `json:"s3_bucket"`
		S3path          string `json:"s3_path"`
	}
	// OutEvent event out from this function to next step
	OutEvent struct {
		DirectMessageEvents []OutDirectMessageEvent `json:"direct-message-events"`
	}
)

var (
	consumerSecret string
	consumerKey    string
	oauthSecret    string
	oauthToken     string
	destBucket     string
	client         *http.Client
	s3srvc         *s3.S3
)

func init() {

	var err error

	consumerKey = os.Getenv("CONSUMER_KEY")
	consumerSecret = os.Getenv("CONSUMER_SECRET_KEY")
	oauthToken = os.Getenv("OAUTH_TOKEN")
	oauthSecret = os.Getenv("OAUTH_SECRET")
	destBucket = os.Getenv("PICTURE_BUCKET")

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

	s3srvc = s3.New(
		session.New(
			&aws.Config{
				Region: aws.String(endpoints.EuNorth1RegionID),
			},
		),
	)

}

// Handler function for the lambda
func Handler(event Event) (OutEvent, error) {

	outDirectMessageEvent := make([]OutDirectMessageEvent, 0)
	for _, v := range event.DirectMessageEvents {

		image, err := getImage(v.MediaURL)
		if err != nil {
			return OutEvent{}, err
		}

		createTime := time.Unix(v.CreateTimestamp/1000, 0)
		s3Prefix := createTime.Format("2006/01/02")
		imageName := fmt.Sprintf("%s.jpg", v.MediaID)
		err = putImageS3(destBucket, s3Prefix, imageName, image)
		if err != nil {
			return OutEvent{}, err
		}

		o := OutDirectMessageEvent{
			CreateTimestamp: v.CreateTimestamp,
			MediaID:         v.MediaID,
			MediaURL:        v.MediaURL,
			URL:             v.URL,
			MessageText:     v.MessageText,
			SenderID:        v.SenderID,
			Text:            v.Text,
			S3bucket:        destBucket,
			S3path:          fmt.Sprintf("%s/%s", s3Prefix, imageName),
		}
		outDirectMessageEvent = append(outDirectMessageEvent, o)
	}

	return OutEvent{
		DirectMessageEvents: outDirectMessageEvent,
	}, nil
}
func putImageS3(bucket string, prefix string, fileName string, image *[]byte) error {

	_, err := s3srvc.PutObject(
		&s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Body:        bytes.NewReader(*image),
			Key:         aws.String(fmt.Sprintf("%s/%s", prefix, fileName)),
			ContentType: aws.String("image/jpeg"),
		},
	)
	if err != nil {
		fmt.Printf("Failed to put object got error: %v\n", err)
		return fmt.Errorf("PUT_IMAGE_S3_FAILED")
	}

	return nil
}
func getImage(URL string) (*[]byte, error) {
	resp, err := client.Get(URL)
	if err != nil {
		fmt.Printf("Failed to get picture fron twitter api. Got error: %v\n", err)
		return nil, fmt.Errorf("FAILED_GET_IMAGE")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("FAILED_READALL_BODY")
	}
	return &body, nil
}

func main() {
	lambda.Start(Handler)
}

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
)

type (
	twitterPayload struct {
		XTwitterWebhooksSignature string `json:"webhooks-signature"`
		RawInput                  string `json:"rawinput"`
		TwitterPayLoad            struct {
			DirectMessageIndicateTypingEvents []struct {
				CreateTimestamp string `json:"created_timestamp"`
				SenderID        string `json:"sender_id"`
				Target          struct {
					RecipientID string `json:"recipient_id"`
				} `json:"target"`
			} `json:"direct_message_indicate_typing_events,omitempty"`
			DirectMessageEvents []struct {
				Type            string `json:"type"`
				ID              string `json:"id"`
				CreateTimestamp string `json:"created_timestamp"`
				MessageCreate   struct {
					SenderID    string `json:"sender_id"`
					MessageData struct {
						Text       string   `json:"text"`
						Entities   struct{} `json:"entities"`
						Attachment struct {
							Type  string `json:"type"`
							Media struct {
								ID         int64  `json:"id"`
								MediaURL   string `json:"media_url"`
								URL        string `json:"url"`
								DisplayURL string `json:"display_url"`
							} `json:"media"`
						} `json:"attachment"`
					} `json:"message_data"`
				} `json:"message_create"`
			} `json:"direct_message_events,omitempty"`
		} `json:"twitter-payload"`
	}
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
)

var (
	consumerSecret string
)

func init() {
	consumerSecret = os.Getenv("CONSUMER_SECRET_KEY")
}

func newEvent(payload twitterPayload) Event {
	directMessageEvents := make([]DirectMessageEvent, 0)
	pictureExists := false
	for _, v := range payload.TwitterPayLoad.DirectMessageEvents {
		if v.MessageCreate.MessageData.Attachment.Media.MediaURL != "" {
			pictureExists = true
		}

		createTime, err := strconv.ParseInt(v.CreateTimestamp, 10, 64)
		if err != nil {
			panic(err)
		}

		d := DirectMessageEvent{
			CreateTimestamp: createTime,
			MediaID:         strconv.FormatInt(v.MessageCreate.MessageData.Attachment.Media.ID, 10),
			URL:             v.MessageCreate.MessageData.Attachment.Media.URL,
			MediaURL:        v.MessageCreate.MessageData.Attachment.Media.MediaURL,
			Text:            v.MessageCreate.MessageData.Text,
			SenderID:        v.MessageCreate.SenderID,
		}
		directMessageEvents = append(directMessageEvents, d)
	}

	return Event{
		DirectMessageEvents: directMessageEvents,
		PictureExists:       pictureExists,
	}
}

// Handler lambda handler.
func Handler(event twitterPayload) (Event, error) {
	body, err := base64.StdEncoding.DecodeString(event.RawInput)
	if err != nil {
		fmt.Printf("DecodeString failed with error: %v\n", err)
	}

	if !verifyRequest(event.XTwitterWebhooksSignature, body) {
		return Event{}, fmt.Errorf("Failed to verify XTwitterWebhooksSignature against body")
	}

	if event.TwitterPayLoad.DirectMessageEvents != nil {
		returnEvent := newEvent(event)
		return returnEvent, nil

	}

	return Event{}, nil
}

func verifyRequest(webhookSignature string, webhookBody []byte) bool {

	crc := webhookSignature
	h := hmac.New(sha256.New, []byte(consumerSecret))
	h.Write(webhookBody)

	crcBase64, err := base64.StdEncoding.DecodeString(crc[7:])
	if err != nil {
		fmt.Printf("verifyRequest failed base64 decodeString with error: %v\n", err)
		return false
	}
	return hmac.Equal(crcBase64, h.Sum(nil))
}

func main() {
	lambda.Start(Handler)
}

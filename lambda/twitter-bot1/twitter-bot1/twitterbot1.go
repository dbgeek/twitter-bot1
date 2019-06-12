package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type (
	twitterPayload struct {
		ForUserID           string `json:"for_user_id"`
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
		} `json:"direct_message_events"`
	}
)

var (
	consumerSecret string
)

type (
	crsToken struct {
		ResponseToken string `json:"response_token"`
	}
)

func init() {
	consumerSecret = os.Getenv("CONSUMER_SECRET_KEY")
}

// Handler main lambda function
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "GET" {
		crcToken := newCrsToken(request.QueryStringParameters["crc_token"])
		respCrcToken, err := json.Marshal(crcToken)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("%s", respCrcToken),
			StatusCode: 200,
		}, nil
	} else if request.HTTPMethod == "POST" {
		if verifyRequest(request) {
			err := postMethod(request)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
				}, nil
			}
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("bad crc\n"),
			StatusCode: 400,
		}, nil

	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func postMethod(request events.APIGatewayProxyRequest) error {
	var payLoad twitterPayload

	err := json.Unmarshal([]byte(request.Body), &payLoad)
	if err != nil {
		return err
	}

	fmt.Printf("payLoad: %v\n", payLoad)

	return nil
}

func newCrsToken(token string) crsToken {
	h := hmac.New(sha256.New, []byte(consumerSecret))
	h.Write([]byte(token))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return crsToken{
		ResponseToken: fmt.Sprintf("sha256=%s", encoded),
	}
}

func verifyRequest(event events.APIGatewayProxyRequest) bool {

	crc := event.Headers["X-Twitter-Webhooks-Signature"]
	h := hmac.New(sha256.New, []byte(consumerSecret))
	h.Write([]byte(event.Body))

	crcBase64, err := base64.StdEncoding.DecodeString(crc[7:])
	if err != nil {
		fmt.Printf("verifyRequest failed base64 decodeString with error: %v\n", err)
		return false
	}
	return hmac.Equal(crcBase64, h.Sum(nil))
}

// String Stringers interface
func (c crsToken) String() string {
	return c.ResponseToken
}

func main() {
	lambda.Start(Handler)
}

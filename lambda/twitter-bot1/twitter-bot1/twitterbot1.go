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
	}
	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("httpMethod: %s", request.HTTPMethod),
		StatusCode: 200,
	}, nil
}

func newCrsToken(token string) crsToken {
	h := hmac.New(sha256.New, []byte(consumerSecret))
	h.Write([]byte(token))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return crsToken{
		ResponseToken: fmt.Sprintf("sha256=%s", encoded),
	}
}

// String Stringers interface
func (c crsToken) String() string {
	return c.ResponseToken
}

func main() {
	lambda.Start(Handler)
}

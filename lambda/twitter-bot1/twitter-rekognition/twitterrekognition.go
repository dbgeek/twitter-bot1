package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/s3"
)

type (
	// Event data struct in
	Event struct {
		DirectMessageEvents []InDirectMessageEvent `json:"direct-message-events"`
		PictureExists       bool                   `json:"picture-exists"`
	}
	// InDirectMessageEvent ..
	InDirectMessageEvent struct {
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
		Faces           []*rekognition.FaceDetail
	}
	// OutEvent event out from this function to next step
	OutEvent struct {
		DirectMessageEvents []OutDirectMessageEvent `json:"direct-message-events"`
	}
)

var (
	s3Svc   *s3.S3
	rekoSvc *rekognition.Rekognition
)

func init() {
	s3Svc = s3.New(session.New(&aws.Config{
		Region: aws.String(endpoints.EuNorth1RegionID),
	}))

	rekoSvc = rekognition.New(session.New(
		&aws.Config{
			Region: aws.String(endpoints.EuWest1RegionID),
		},
	))
}

// Handler lambda handler function
func Handler(events Event) (OutEvent, error) {
	outDirectMessageEvent := make([]OutDirectMessageEvent, 0)

	for _, event := range events.DirectMessageEvents {
		fmt.Println("*****START PROCESSING EVENT*****")
		picture, err := getImageS3(event.S3bucket, event.S3path)
		if err != nil {
			fmt.Printf("Failed to get picture from s3. Got error: %v\n", err)
		}

		faceDetails, err := detectFaces(picture)

		o := OutDirectMessageEvent{
			CreateTimestamp: event.CreateTimestamp,
			MediaID:         event.MediaID,
			MediaURL:        event.MediaURL,
			URL:             event.URL,
			MessageText:     event.MessageText,
			SenderID:        event.SenderID,
			Text:            event.Text,
			S3bucket:        event.S3bucket,
			S3path:          event.S3path,
			Faces:           faceDetails,
		}

		buffOfFaceDetails, err := json.Marshal(faceDetails)
		if err != nil {
			fmt.Printf("Marshal facedetails failed with error: %v \n", err)
		}

		_, err = s3Svc.PutObject(
			&s3.PutObjectInput{
				Bucket:      aws.String(event.S3bucket),
				Body:        bytes.NewReader(buffOfFaceDetails),
				Key:         aws.String(fmt.Sprintf("%s.json", event.S3path)),
				ContentType: aws.String("text/plain"),
			},
		)
		if err != nil {
			fmt.Printf("Failed to put object got error: %v\n", err)
		}

		outDirectMessageEvent = append(outDirectMessageEvent, o)

	}
	return OutEvent{
		DirectMessageEvents: outDirectMessageEvent,
	}, nil

}

func getImageS3(bucket string, key string) (*[]byte, error) {
	data, err := s3Svc.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		fmt.Printf("Failed to get picture from s3. Got error: %v\n", err)
	}
	defer data.Body.Close()
	fmt.Printf("Contentlength: %v etag: %v\n", *data.ContentLength, *data.ETag)
	picture, err := ioutil.ReadAll(data.Body)
	if err != nil {
		fmt.Printf("Failed to read data.Body. Got error: %v\n", err)
	}
	return &picture, nil
}

func detectFaces(picture *[]byte) ([]*rekognition.FaceDetail, error) {
	input := &rekognition.DetectFacesInput{
		Attributes: []*string{aws.String("ALL")},
		Image: &rekognition.Image{
			Bytes: *picture,
		},
	}

	result, err := rekoSvc.DetectFaces(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rekognition.ErrCodeInvalidS3ObjectException:
				fmt.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
			case rekognition.ErrCodeInvalidParameterException:
				fmt.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
			case rekognition.ErrCodeImageTooLargeException:
				fmt.Println(rekognition.ErrCodeImageTooLargeException, aerr.Error())
			case rekognition.ErrCodeAccessDeniedException:
				fmt.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
			case rekognition.ErrCodeInternalServerError:
				fmt.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
			case rekognition.ErrCodeThrottlingException:
				fmt.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
			case rekognition.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case rekognition.ErrCodeInvalidImageFormatException:
				fmt.Println(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}
	return result.FaceDetails, nil
}

func main() {
	lambda.Start(Handler)
}

package golambda

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type (
	Message struct {
		Records []Record
	}

	Record struct {
		EventVersion         string
		EventSubscriptionArn string
		EventSource          string
		Sns                  interface{} `json:",omitempty"`
	}
)


func NewLambda(endpoint string) (*lambda.Lambda, error) {
	creds := credentials.NewStaticCredentials("id", "secret", "token")

	awsConfig := aws.NewConfig().
		WithRegion("faux-region-1").
		WithEndpoint(endpoint).
		WithCredentials(creds)

	s, err := session.NewSession(awsConfig)

	if err != nil {
		return nil, err
	}

	return lambda.New(s), nil
}

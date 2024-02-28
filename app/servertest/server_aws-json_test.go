package servertest

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSQS_json(t *testing.T, region string, endpoint string) *sqs.Client {
	creds := credentials.NewStaticCredentialsProvider("id", "secret", "token")

	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	require.NoError(t, err, "Unable to load SDK")

	return sqs.NewFromConfig(sdkConfig, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.Credentials = creds
		o.Region = region
	})
}

func noOpV2(*sqs.Client, *string) error {
	return nil
}

func TestNewIntegration_Json(t *testing.T) {
	testTable := []struct {
		Name      string
		Expected  []string
		QueueFunc func(*sqs.Client, *string) error
	}{
		{
			Name:      "Empty queue OK",
			Expected:  []string{},
			QueueFunc: noOpV2,
		},
	}
	for _, tr := range testTable {
		t.Run(tr.Name, func(t *testing.T) {
			// Start local SQS
			srv, err := New("")
			noSetupError(t, err)
			defer srv.Quit()

			sqsClient := newSQS_json(t, "faux-region-1", srv.URL())

			// Create test queue
			_, err = sqsClient.CreateQueue(
				context.TODO(),
				&sqs.CreateQueueInput{QueueName: aws.String("test-queue")})
			noSetupError(t, err)

			getQueueUrlOutput, err := sqsClient.GetQueueUrl(
				context.TODO(),
				&sqs.GetQueueUrlInput{QueueName: aws.String("test-queue")})
			noSetupError(t, err)
			queueURL := getQueueUrlOutput.QueueUrl

			// Setup Queue Sate
			err = tr.QueueFunc(sqsClient, queueURL)
			noSetupError(t, err)

			// Test
			receiveMessageInput := &sqs.ReceiveMessageInput{QueueUrl: queueURL}
			receiveMessageOutput, err := sqsClient.ReceiveMessage(
				context.TODO(),
				receiveMessageInput)

			msgsBody := []string{}
			for _, b := range receiveMessageOutput.Messages {
				msgsBody = append(msgsBody, *b.Body)
			}

			assert.Equal(t, tr.Expected, msgsBody, "Messages")
			assert.Equal(t, nil, err, "Error")
		})
	}
}

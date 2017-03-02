package servertest_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/p4tin/goaws/app/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Consume address
	srv, err := servertest.New("localhost:4100")
	noSetupError(t, err)
	defer srv.Quit()

	// Test
	_, err = servertest.New("localhost:4100")
	assert.Equal(t, errors.New("cannot listen on localhost: listen tcp 127.0.0.1:4100: bind: address already in use"), err, "Error")
}

func TestNewIntegration(t *testing.T) {
	testTable := []struct {
		Name      string
		Expected  []string
		QueueFunc func(sqsiface.SQSAPI, *string) error
	}{
		{
			Name:      "Empty queue OK",
			Expected:  []string{},
			QueueFunc: noOp,
		},
		{
			Name:     "Some messages OK",
			Expected: []string{"hello world"},
			QueueFunc: func(svc sqsiface.SQSAPI, queueURL *string) error {
				_, err := svc.SendMessage(&sqs.SendMessageInput{
					MessageBody: aws.String("hello world"),
					QueueUrl:    queueURL,
				})
				return err
			},
		},
	}
	for _, tr := range testTable {
		t.Run(tr.Name, func(t *testing.T) {
			// Start local SQS
			srv, err := servertest.New("")
			noSetupError(t, err)
			defer srv.Quit()

			svc := newSQS(t, "faux-region-1", srv.URL())

			// Create test queue
			_, err = svc.CreateQueue(
				&sqs.CreateQueueInput{QueueName: aws.String("test-queue")})
			noSetupError(t, err)

			getQueueUrlOutput, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: aws.String("test-queue")})
			noSetupError(t, err)
			queueURL := getQueueUrlOutput.QueueUrl

			// Setup Queue Sate
			err = tr.QueueFunc(svc, queueURL)
			noSetupError(t, err)

			// Test
			receiveMessageInput := &sqs.ReceiveMessageInput{QueueUrl: queueURL}
			receiveMessageOutput, err := svc.ReceiveMessage(receiveMessageInput)

			msgsBody := []string{}
			for _, b := range receiveMessageOutput.Messages {
				msgsBody = append(msgsBody, *b.Body)
			}

			assert.Equal(t, tr.Expected, msgsBody, "Messages")
			assert.Equal(t, nil, err, "Error")
		})
	}
}

func newSQS(t *testing.T, region string, endpoint string) *sqs.SQS {
	creds := credentials.NewStaticCredentials("id", "secret", "token")

	awsConfig := aws.NewConfig().
		WithRegion(region).
		WithEndpoint(endpoint).
		WithCredentials(creds)


	session1 := session.New(awsConfig)

	svc := sqs.New(session1)
	return svc
}

func noOp(sqsiface.SQSAPI, *string) error {
	return nil
}

func noSetupError(t *testing.T, err error) {
	require.NoError(t, err, "Failed to setup for test")
}

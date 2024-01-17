package servertest

import (
	"errors"
	"testing"

	"github.com/Admiral-Piett/goaws/app/utils"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Admiral-Piett/goaws/app"
	"github.com/Admiral-Piett/goaws/app/router"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	utils.InitializeDecoders()
	m.Run()
}

func TestNew(t *testing.T) {
	// Consume address
	srv, err := New("localhost:4100")
	noSetupError(t, err)
	defer srv.Quit()

	// Test
	_, err = New("localhost:4100")
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
		//{
		//	Name:     "Some messages OK",
		//	Expected: []string{"hello world"},
		//	QueueFunc: func(svc sqsiface.SQSAPI, queueURL *string) error {
		//		attributes := make(map[string]*sqs.MessageAttributeValue)
		//		attributes["some string"] = &sqs.MessageAttributeValue{
		//			StringValue: aws.String("string value with a special character \u2318"),
		//			DataType:    aws.String("String"),
		//		}
		//		attributes["some number"] = &sqs.MessageAttributeValue{
		//			StringValue: aws.String("123"),
		//			DataType:    aws.String("Number"),
		//		}
		//		attributes["some binary"] = &sqs.MessageAttributeValue{
		//			BinaryValue: []byte{1, 2, 3},
		//			DataType:    aws.String("Binary"),
		//		}
		//
		//		response, err := svc.SendMessage(&sqs.SendMessageInput{
		//			MessageBody:       aws.String("hello world"),
		//			MessageAttributes: attributes,
		//			QueueUrl:          queueURL,
		//		})
		//
		//		assert.Equal(t, "5eb63bbbe01eeed093cb22bb8f5acdc3", *response.MD5OfMessageBody)
		//		assert.Equal(t, "7820c7a3712c7c359cf80485f67aa34d", *response.MD5OfMessageAttributes)
		//		return err
		//	},
		//},
	}
	for _, tr := range testTable {
		t.Run(tr.Name, func(t *testing.T) {
			// Start local SQS
			srv, err := New("")
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

func TestSNSRoutes(t *testing.T) {
	// Consume address
	srv, err := NewSNSTest("localhost:4100", &snsTest{t: t})

	noSetupError(t, err)
	defer srv.Quit()

	creds := credentials.NewStaticCredentials("id", "secret", "token")

	awsConfig := aws.NewConfig().
		WithRegion("us-east-1").
		WithEndpoint(srv.URL()).
		WithCredentials(creds)

	session1 := session.New(awsConfig)
	client := sns.New(session1)

	response, err := client.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String("testing"),
	})
	require.NoError(t, err, "SNS Create Topic Failed")

	params := &sns.SubscribeInput{
		Protocol: aws.String("sqs"), // Required
		TopicArn: response.TopicArn, // Required
		Endpoint: aws.String(srv.URL() + "/local-sns"),
	}
	subscribeResponse, err := client.Subscribe(params)
	require.NoError(t, err, "SNS Subscribe Failed")
	t.Logf("Succesfully subscribed: %s\n", *subscribeResponse.SubscriptionArn)

	publishParams := &sns.PublishInput{
		Message:  aws.String("Cool"),
		TopicArn: response.TopicArn,
	}
	publishResponse, err := client.Publish(publishParams)
	require.NoError(t, err, "SNS Publish Failed")
	t.Logf("Succesfully published: %s\n", *publishResponse.MessageId)
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

type snsTest struct {
	t *testing.T
}

func NewSNSTest(addr string, snsTest *snsTest) (*Server, error) {
	if addr == "" {
		addr = "localhost:0"
	}
	localURL := strings.Split(addr, ":")
	app.CurrentEnvironment.Host = localURL[0]
	app.CurrentEnvironment.Port = localURL[1]
	log.WithFields(log.Fields{
		"host": app.CurrentEnvironment.Host,
		"port": app.CurrentEnvironment.Port,
	}).Info("URL Starting to listen")

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("cannot listen on localhost: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot listen on localhost: %v", err)
	}

	r := mux.NewRouter()
	r.Handle("/", router.New())
	snsTest.SetSNSRoutes("/local-sns", r, nil)

	srv := Server{listener: l, handler: r}

	go http.Serve(l, &srv)

	return &srv, nil
}

// Define handlers for various AWS SNS POST calls
func (s *snsTest) SetSNSRoutes(urlPath string, r *mux.Router, handler http.Handler) {

	r.HandleFunc(urlPath, s.SubscribeConfirmHandle).Methods("POST").Headers("x-amz-sns-message-type", "SubscriptionConfirmation")
	if handler != nil {
		log.WithFields(log.Fields{
			"urlPath": urlPath,
		}).Debug("handler not nil")
		// handler is supposed to be wrapper that inturn calls NotificationHandle
		r.Handle(urlPath, handler).Methods("POST").Headers("x-amz-sns-message-type", "Notification")
	} else {
		log.WithFields(log.Fields{
			"urlPath": urlPath,
		}).Debug("handler nil")
		// if no wrapper handler available then define anonymous handler and directly call NotificationHandle
		r.HandleFunc(urlPath, func(rw http.ResponseWriter, req *http.Request) {
			s.NotificationHandle(rw, req)
		}).Methods("POST").Headers("x-amz-sns-message-type", "Notification")
	}
}

func (s *snsTest) SubscribeConfirmHandle(rw http.ResponseWriter, req *http.Request) {
	//params := &sns.ConfirmSubscriptionInput{
	//	Token:    aws.String(msg.Token),    // Required
	//	TopicArn: aws.String(msg.TopicArn), // Required
	//}
	var f interface{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		s.t.Log("Unable to Parse Body")
	}
	s.t.Log(string(body))
	err = json.Unmarshal(body, &f)
	if err != nil {
		s.t.Log("Unable to Unmarshal request")
	}

	data := f.(map[string]interface{})
	s.t.Log(data["Type"].(string))

	if data["Type"].(string) == "SubscriptionConfirmation" {
		subscribeURL := data["SubscribeURL"].(string)
		time.Sleep(time.Second)
		response, err := http.Get(subscribeURL)
		if err != nil {
			s.t.Logf("Unable to confirm subscriptions. %s\n", err)
			s.t.Fail()
		} else {
			s.t.Logf("Subscription Confirmed successfully. %d\n", response.StatusCode)
		}
	} else if data["Type"].(string) == "Notification" {
		s.t.Log("Received this message : ", data["Message"].(string))
	}
}

func (s *snsTest) NotificationHandle(rw http.ResponseWriter, req *http.Request) []byte {
	subArn := req.Header.Get("X-Amz-Sns-Subscription-Arn")

	msg := app.SNSMessage{}
	_, err := DecodeJSONMessage(req, &msg)
	if err != nil {
		log.Error(err)
		return []byte{}
	}

	s.t.Logf("NotificationHandle %s  MSG(%s)", subArn, msg.Message)
	return []byte(msg.Message)
}

func DecodeJSONMessage(req *http.Request, v interface{}) ([]byte, error) {

	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, errors.New("empty payload")
	}
	err = json.Unmarshal([]byte(payload), v)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

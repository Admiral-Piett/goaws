package models

import "net/http"

func init() {
	SqsErrors = map[string]SqsErrorType{
		"QueueNotFound":                {HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleQueueService.NonExistentQueue", Message: "The specified queue does not exist for this wsdl version."},
		"QueueExists":                  {HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue already exists."},
		"MessageDoesNotExist":          {HttpError: http.StatusNotFound, Type: "Not Found", Code: "AWS.SimpleQueueService.QueueExists", Message: "The specified queue does not contain the message specified."},
		"GeneralError":                 {HttpError: http.StatusBadRequest, Type: "GeneralError", Code: "AWS.SimpleQueueService.GeneralError", Message: "General Error."},
		"TooManyEntriesInBatchRequest": {HttpError: http.StatusBadRequest, Type: "TooManyEntriesInBatchRequest", Code: "AWS.SimpleQueueService.TooManyEntriesInBatchRequest", Message: "Maximum number of entries per request are 10."},
		"BatchEntryIdsNotDistinct":     {HttpError: http.StatusBadRequest, Type: "BatchEntryIdsNotDistinct", Code: "AWS.SimpleQueueService.BatchEntryIdsNotDistinct", Message: "Two or more batch entries in the request have the same Id."},
		"EmptyBatchRequest":            {HttpError: http.StatusBadRequest, Type: "EmptyBatchRequest", Code: "AWS.SimpleQueueService.EmptyBatchRequest", Message: "The batch request doesn't contain any entries."},
		"InvalidVisibilityTimeout":     {HttpError: http.StatusBadRequest, Type: "ValidationError", Code: "AWS.SimpleQueueService.ValidationError", Message: "The visibility timeout is incorrect"},
		"MessageNotInFlight":           {HttpError: http.StatusBadRequest, Type: "MessageNotInFlight", Code: "AWS.SimpleQueueService.MessageNotInFlight", Message: "The message referred to isn't in flight."},
		"MessageTooBig":                {HttpError: http.StatusBadRequest, Type: "MessageTooBig", Code: "InvalidParameterValue", Message: "The message size exceeds the limit."},
		"InvalidParameterValue":        {HttpError: http.StatusBadRequest, Type: "InvalidParameterValue", Code: "AWS.SimpleQueueService.InvalidParameterValue", Message: "An invalid or out-of-range value was supplied for the input parameter."},
		"InvalidAttributeValue":        {HttpError: http.StatusBadRequest, Type: "InvalidAttributeValue", Code: "AWS.SimpleQueueService.InvalidAttributeValue", Message: "Invalid Value for the parameter RedrivePolicy."},
	}
	SnsErrors = map[string]SnsErrorType{
		"InvalidParameterValue": {HttpError: http.StatusBadRequest, Type: "InvalidParameterValue", Code: "AWS.SimpleNotificationService.InvalidParameterValue", Message: "An invalid or out-of-range value was supplied for the input parameter."},
		"TopicNotFound":         {HttpError: http.StatusBadRequest, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentTopic", Message: "The specified topic does not exist for this wsdl version."},
		"SubscriptionNotFound":  {HttpError: http.StatusNotFound, Type: "Not Found", Code: "AWS.SimpleNotificationService.NonExistentSubscription", Message: "The specified subscription does not exist for this wsdl version."},
		"TopicExists":           {HttpError: http.StatusBadRequest, Type: "Duplicate", Code: "AWS.SimpleNotificationService.TopicAlreadyExists", Message: "The specified topic already exists."},
		"ValidationError":       {HttpError: http.StatusBadRequest, Type: "InvalidParameter", Code: "AWS.SimpleNotificationService.ValidationError", Message: "The input fails to satisfy the constraints specified by an AWS service."},
	}
}

type SqsErrorType struct {
	HttpError int
	Type      string
	Code      string
	Message   string
}

func (s SqsErrorType) StatusCode() int {
	return s.HttpError
}

func (s SqsErrorType) Response() ErrorResult {
	return ErrorResult{Type: s.Type, Code: s.Code, Message: s.Message}
}

var SqsErrors map[string]SqsErrorType

type SnsErrorType struct {
	HttpError int
	Type      string
	Code      string
	Message   string
}

func (s SnsErrorType) StatusCode() int {
	return s.HttpError
}

func (s SnsErrorType) Response() ErrorResult {
	return ErrorResult{Type: s.Type, Code: s.Code, Message: s.Message}
}

var SnsErrors map[string]SnsErrorType

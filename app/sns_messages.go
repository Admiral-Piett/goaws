package app

/*** List Topics Response */
type TopicArnResult struct {
	TopicArn string `xml:"TopicArn"`
}

type TopicNamestype struct {
	Member []TopicArnResult `xml:"member"`
}

type ListTopicsResult struct {
	Topics TopicNamestype `xml:"Topics"`
}

type ListTopicsResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Result   ListTopicsResult `xml:"ListTopicsResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** Create Topic Response */
type CreateTopicResult struct {
	TopicArn string `xml:"TopicArn"`
}

type CreateTopicResponse struct {
	Xmlns    string            `xml:"xmlns,attr"`
	Result   CreateTopicResult `xml:"CreateTopicResult"`
	Metadata ResponseMetadata  `xml:"ResponseMetadata"`
}

/*** Create Subscription ***/
type SubscribeResult struct {
	SubscriptionArn string `xml:"SubscriptionArn"`
}

type SubscribeResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Result   SubscribeResult  `xml:"SubscribeResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** ConfirmSubscriptionResponse ***/
type ConfirmSubscriptionResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Result   SubscribeResult  `xml:"ConfirmSubscriptionResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/***  Set Subscription Response ***/

type SetSubscriptionAttributesResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** Get Subscription Attributes ***/
type GetSubscriptionAttributesResult struct {
	SubscriptionAttributes SubscriptionAttributes `xml:"Attributes,omitempty"`
}

type GetTopicAttributesResult struct {
	Attributes *TopicAttributes `xml:"Attributes,omitempty"`
}

type SubscriptionAttributes struct {
	/* SubscriptionArn, FilterPolicy */
	Entries *[]SubscriptionAttributeEntry `xml:"entry,omitempty"`
}

type TopicAttributes struct {
	/* TopicArn, FilterPolicy */
	Entries *[]TopicAttributeEntry `xml:"entry,omitempty"`
}

type SubscriptionAttributeEntry struct {
	Key   string `xml:"key,omitempty"`
	Value string `xml:"value,omitempty"`
}

type TopicAttributeEntry struct {
	Key   string `xml:"key,omitempty"`
	Value string `xml:"value,omitempty"`
}

type GetSubscriptionAttributesResponse struct {
	Xmlns    string                          `xml:"xmlns,attr,omitempty"`
	Result   GetSubscriptionAttributesResult `xml:"GetSubscriptionAttributesResult"`
	Metadata ResponseMetadata                `xml:"ResponseMetadata,omitempty"`
}

type GetTopicAttributesResponse struct {
	Xmlns    string                   `xml:"xmlns,attr,omitempty"`
	Result   GetTopicAttributesResult `xml:"GetTopicAttributesResult" binding:"required"`
	Metadata *ResponseMetadata        `xml:"ResponseMetadata,omitempty"`
}

/*** List Subscriptions Response */
type TopicMemberResult struct {
	TopicArn        string  `xml:"TopicArn" binding:"required"`
	Protocol        *string `xml:"Protocol,omitempty"`
	SubscriptionArn *string `xml:"SubscriptionArn,omitempty"`
	Owner           string  `xml:"Owner" binding:"required"`
	Endpoint        *string `xml:"Endpoint,omitempty"`
}

type TopicSubscriptions struct {
	Member []TopicMemberResult `xml:"member"`
}

type ListSubscriptionsResult struct {
	Subscriptions TopicSubscriptions `xml:"Subscriptions"`
}

type ListSubscriptionsResponse struct {
	Xmlns    string                  `xml:"xmlns,attr"`
	Result   ListSubscriptionsResult `xml:"ListSubscriptionsResult"`
	Metadata ResponseMetadata        `xml:"ResponseMetadata"`
}

/*** List Subscriptions By Topic Response */

type ListSubscriptionsByTopicResult struct {
	Subscriptions TopicSubscriptions `xml:"Subscriptions"`
}

type ListSubscriptionsByTopicResponse struct {
	Xmlns    string                  `xml:"xmlns,attr"`
	Result   ListSubscriptionsResult `xml:"ListSubscriptionsResult"`
	Metadata ResponseMetadata        `xml:"ResponseMetadata"`
}

/*** Publish ***/

type PublishResult struct {
	MessageId string `xml:"MessageId"`
}

type PublishResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Result   PublishResult    `xml:"PublishResult"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** Unsubscribe ***/
type UnsubscribeResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** Delete Topic ***/
type DeleteTopicResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** Policy ***/
type TopicAttributePolicy struct {
	Version   string          `json:"Version" binding:"required"`
	Id        string          `json:"Id" binding:"required"`
	Statement *[]AWSStatement `json:"Statement,omitempty"`
}

type AWSStatement struct {
	Effect    *string                 `json:"Effect,omitempty"`
	Sid       string                  `json:"Sid" binding:"required"`
	Principal AWSPrincipal            `json:"Principal" binding:"required"`
	Action    []string                `json:"Action" binding:"required"`
	Resource  string                  `json:"Resource" binding:"required"`
	Condition *map[string]interface{} `json:"Condition,omitempty"`
}

type AWSPrincipal struct {
	AWS string `json:"AWS" binding:"required"`
}

type DeliveryPolicy struct {
	DefaultHealthyRetryPolicy *RetryPolicy `json:"defaultHealthyRetryPolicy,omitempty"`
	SicklyRetryPolicy         *RetryPolicy `json:"sicklyRetryPolicy,omitempty"`
	ThrottlePolicy            *RetryPolicy `json:"throttlePolicy,omitempty"`
	Guaranteed                *bool        `json:"guaranteed,omitempty"`
}

type RetryPolicy struct {
	NumberNoDelayRetries     *int    `json:"numNoDelayRetries,omitempty"`
	NumberMinDelayRetries    *int    `json:"numMinDelayRetries,omitempty"`
	MinimumDelayTarget       *int    `json:"minDelayTarget,omitempty"`
	MaximumDelayTarget       *int    `json:"maxDelayTarget,omitempty"`
	NumberMaxDelayRetries    *int    `json:"numMaxDelayRetries,omitempty"`
	NumberRetries            *int    `json:"numRetries,omitempty"`
	BackoffFunction          *string `json:"backoffFunction,omitempty"`
	MaximumReceivesPerSecond *int    `json:"maxReceivesPerSecond,omitempty"`
}

package app

/***  Set Subscription Response ***/
type SetSubscriptionAttributesResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** List Subscriptions By Topic Response */

type TopicMemberResult struct {
	TopicArn        string `xml:"TopicArn"`
	Protocol        string `xml:"Protocol"`
	SubscriptionArn string `xml:"SubscriptionArn"`
	Owner           string `xml:"Owner"`
	Endpoint        string `xml:"Endpoint"`
}

type TopicSubscriptions struct {
	Member []TopicMemberResult `xml:"member"`
}

type ListSubscriptionsByTopicResult struct {
	Subscriptions TopicSubscriptions `xml:"Subscriptions"`
}

type ListSubscriptionsByTopicResponse struct {
	Xmlns    string                         `xml:"xmlns,attr"`
	Result   ListSubscriptionsByTopicResult `xml:"ListSubscriptionsByTopicResult"`
	Metadata ResponseMetadata               `xml:"ResponseMetadata"`
}

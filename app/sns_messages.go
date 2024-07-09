package app

/***  Set Subscription Response ***/
type SetSubscriptionAttributesResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

/*** Get Subscription Attributes ***/
type GetSubscriptionAttributesResult struct {
	SubscriptionAttributes SubscriptionAttributes `xml:"Attributes,omitempty"`
}

type SubscriptionAttributes struct {
	/* SubscriptionArn, FilterPolicy */
	Entries []SubscriptionAttributeEntry `xml:"entry,omitempty"`
}

type SubscriptionAttributeEntry struct {
	Key   string `xml:"key,omitempty"`
	Value string `xml:"value,omitempty"`
}

type GetSubscriptionAttributesResponse struct {
	Xmlns    string                          `xml:"xmlns,attr,omitempty"`
	Result   GetSubscriptionAttributesResult `xml:"GetSubscriptionAttributesResult"`
	Metadata ResponseMetadata                `xml:"ResponseMetadata,omitempty"`
}

/*** List Subscriptions Response */
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
	Xmlns    string                         `xml:"xmlns,attr"`
	Result   ListSubscriptionsByTopicResult `xml:"ListSubscriptionsByTopicResult"`
	Metadata ResponseMetadata               `xml:"ResponseMetadata"`
}

/*** Delete Topic ***/
type DeleteTopicResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

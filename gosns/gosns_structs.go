package gosns

/*** Common ***/
type ResponseMetadata struct {
	RequestId string                `xml:"RequestId"`
}


/*** List Topics Response */
type TopicArnResult struct {
	TopicArn string				`xml:"TopicArn"`
}

type TopicNamestype struct {
	Member []TopicArnResult			`xml:"member"`
}

type  ListTopicsResult struct {
	Topics TopicNamestype                        `xml:"Topics"`
}

type ListTopicsResponse struct {
	Xmlns  		string  		`xml:"xmlns,attr"`
	Result		ListTopicsResult	`xml:"ListTopicsResult"`
	Metadata 	ResponseMetadata	`xml:"ResponseMetadata"`
}

/*** Create Topic Response */
type CreateTopicResult struct {
	TopicArn string			`xml:"TopicArn"`
}

type CreateTopicResponse struct {
	Xmlns 		string			`xml:"xmlns,attr"`
	Result		CreateTopicResult	`xml:"CreateTopicResult"`
	Metadata 	ResponseMetadata	`xml:"ResponseMetadata"`
}


/*** Create Subscription ***/
type SubscribeResult struct {
	SubscriptionArn string			`xml:"SubscriptionArn"`
}

type SubscribeResponse struct {
	Xmlns 		string			`xml:"xmlns,attr"`
	Result		SubscribeResult		`xml:"SubscribeResult"`
	Metadata 	ResponseMetadata	`xml:"ResponseMetadata"`
}


/***  Set Subscription Response ***/

type SetSubscriptionAttributesResponse struct {
	Xmlns  		string  		`xml:"xmlns,attr"`
	Metadata 	ResponseMetadata	`xml:"ResponseMetadata"`

}

/*** List Subscriptions Response */
type TopicMemberResult struct {
	TopicArn 		string		`xml:"TopicArn"`
	Protocol		string		`xml:"Protocol"`
	SubscriptionArn		string		`xml:"SubscriptionArn"`
	Owner			string		`xml:"Owner"`
	Endpoint 		string		`xml:"Endpoint"`
}

type TopicSubscriptions struct {
	Member []TopicMemberResult			`xml:"member"`
}

type  ListSubscriptionsResult struct {
	Subscriptions TopicSubscriptions               `xml:"Subscriptions"`
}

type ListSubscriptionsResponse struct {
	Xmlns  		string  		`xml:"xmlns,attr"`
	Result		ListSubscriptionsResult	`xml:"ListSubscriptionsResult"`
	Metadata 	ResponseMetadata	`xml:"ResponseMetadata"`
}



/*** Publish ***/

type PublishResult struct {
	MessageId string			`xml:"MessageId"`
}

type PublishResponse struct {
	Xmlns 		string			`xml:"xmlns,attr"`
	Result		PublishResult		`xml:"PublishResult"`
	Metadata 	ResponseMetadata	`xml:"ResponseMetadata"`

}

/*** Error Responses ***/
type ErrorResult struct {
	Type string 		`xml:"Type,omitempty"`
	Code string		`xml:"Code,omitempty"`
	Message string		`xml:"Message,omitempty"`
	RequestId string	`xml:"RequestId,omitempty"`
}

type ErrorResponse struct {
	Result	ErrorResult	`xml:"Error"`
}

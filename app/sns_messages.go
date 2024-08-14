package app

/***  Set Subscription Response ***/
type SetSubscriptionAttributesResponse struct {
	Xmlns    string           `xml:"xmlns,attr"`
	Metadata ResponseMetadata `xml:"ResponseMetadata"`
}

package app

/*** Common ***/
type ResponseMetadata struct {
	RequestId string `xml:"RequestId"`
}


/*** Error Responses ***/
type ErrorResult struct {
	Type      string `xml:"Type,omitempty"`
	Code      string `xml:"Code,omitempty"`
	Message   string `xml:"Message,omitempty"`
	RequestId string `xml:"RequestId,omitempty"`
}

type ErrorResponse struct {
	Result ErrorResult `xml:"Error"`
}





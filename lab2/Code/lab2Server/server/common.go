package server

import (
	"encoding/xml"
)

// Since xml tag name and field name are the same,
// there is no need in tags.
// But i'd better be sure.


type request struct {
	XMLName xml.Name `xml:"Request"`
	ClientName string `xml:"ClientName"`
}

type response struct {
	XMLName xml.Name `xml:"Response"`
	Clients []*client `xml:"Client"`
}

// I return time as unix timestamp.
// If it's necessary to pass real date like "2020\etc",
// then it could be rewritten in matter of 5-10 minutes.
type client struct {
	XMLName xml.Name `xml:"Client"`
	ClientName string `xml:"ClientName"`
	ClientTime int64 `xml:"ClientTime"`
	TimerTime int64 `xml:"TimerTime"`
}

// Was called only once, so i decided to use simple anonymous function

//// errorHandler - handles possible error leak.
//// Used in defer statements with error
//func errorHandler(f func() error) {
//	if err := f(); err != nil {
//		log.Println("ERROR: ", err)
//	}
//}
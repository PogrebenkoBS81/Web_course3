package client

import (
	"encoding/xml"
)

// Since xml tag name and field name are the same,
// there is no need in tags.
// But i'd better be sure.

// I return time as unix timestamp.
// If it's necessary to pass real date like "2020\etc",
// then it could be rewritten in matter of 5-10 minutes.

const (
	maxBufioSize = 4096
)

type request struct {
	XMLName    xml.Name `xml:"Request"`
	ClientName string   `xml:"ClientName"`
}

type response struct {
	XMLName xml.Name  `xml:"Response"`
	Clients []*client `xml:"Client"`
	Timer   int64     `xml:"Timer"` // Time when timer was started
}

type client struct {
	XMLName   xml.Name `xml:"Client"`
	Connected int64    `xml:"ClientTime"` // time when client connected
	Name      string   `xml:"ClientName"` // client name
	IP        string   `xml:"ClientIP"`   // client IP (to distinguish clients with the same names)
}

// Was called only once, so i decided to use simple anonymous function

////errorHandler - handles possible error leak.
////Used in defer statements with error
//func errorHandler(f func() error) {
//	if err := f(); err != nil {
//		log.Println("ERROR: ", err)
//	}
//}

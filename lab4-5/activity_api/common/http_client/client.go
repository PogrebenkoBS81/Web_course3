package http_client

import (
	"activity_api/common/error_manage"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// statusCodes - OK status codes
var statusCodes = map[int]bool{
	http.StatusAccepted: true,
	http.StatusCreated:  true,
	http.StatusOK:       true,
}

// NetHTTP - simple http client
type NetHTTP struct {
	client *http.Client
}

// NewHTTPClient - returns new simple http client
func NewHTTPClient() *NetHTTP {
	return &NetHTTP{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: time.Millisecond * 1500,
		},
	}
}

// setHeaders - sets request headers
func (n *NetHTTP) setHeaders(req *http.Request, headers map[string]string) {
	for header, value := range headers {
		req.Header.Set(header, value)
	}
}

// MakeRequest - makes request to given handler with given parameters
func (n *NetHTTP) MakeRequest(method, path string, headers map[string]string, body []byte) (bts []byte, err error) {
	var req *http.Request

	if body != nil {
		req, err = http.NewRequest(method, path, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}

	if err != nil {
		return nil, err
	}

	n.setHeaders(req, headers)

	res, err := n.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() { n.closeResponse(res) }()

	if code := res.StatusCode; !statusCodes[code] {
		bts, _ := ioutil.ReadAll(res.Body)

		return nil, fmt.Errorf("error status code: %d, resp: %s", code, string(bts))
	}

	return ioutil.ReadAll(res.Body)
}

// closeResponse - response close helpers.
func (n *NetHTTP) closeResponse(res *http.Response) {
	_, err := io.Copy(ioutil.Discard, res.Body)
	// Usual Log() is used here, because this client won't be used in AAService
	error_manage.ErrorWriter(err)

	error_manage.ErrorWriter(res.Body.Close())
}

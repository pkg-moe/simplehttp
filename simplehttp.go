package simplehttp

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	dialContext = (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext

	defaultTransport = &http.Transport{
		DialContext:       dialContext,
		DisableKeepAlives: true,
	}
)

// NewClient create a http client
func NewClient() *http.Client {
	return &http.Client{Transport: defaultTransport, Timeout: 5 * time.Second}
}

// Behaves as https://golang.org/pkg/net/http/#Client.Do with the exception that
// the Response.Body does not need to be closed. This function should generally
// only be used when it is already known that the response body will be
// relatively small, as it will be completely read into memory.
func Do(client *http.Client, req *http.Request) (*Response, error) {
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bb := &bytes.Buffer{}
	n, err := io.Copy(bb, resp.Body)
	if err != nil {
		return nil, err
	}

	resp.ContentLength = n
	resp.Body = io.NopCloser(bb)

	return &Response{Response: *resp}, nil
}

// Behaves as https://golang.org/pkg/net/http/#Get but uses simplehttp.Do() to
// make the request.
func Get(client *http.Client, url string) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return Do(client, req)
}

// Behaves as https://golang.org/pkg/net/http/#Head but uses simplehttp.Do() to
// make the request.
func Head(client *http.Client, url string) (*Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return Do(client, req)
}

// Behaves as https://golang.org/pkg/net/http/#Post but uses simplehttp.Do() to
// make the request.
func Post(client *http.Client, url string, contentType string, body io.Reader) (*Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return Do(client, req)
}

// Behaves as https://golang.org/pkg/net/http/#PostForm but uses simplehttp.Do()
// to make the request.
func PostForm(client *http.Client, url string, data url.Values) (*Response, error) {
	return Post(client, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

type Response struct {
	http.Response
}

func (r *Response) Bytes() ([]byte, error) {
	b, err := io.ReadAll(r.Body)
	_ = r.Body.Close()
	return b, err
}

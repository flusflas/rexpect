package rexpect

import (
	"net/http"
	"net/url"
	"strings"
)

// RequestBodyMatchFunc represents a function to perform a custom match on the request body.
type RequestBodyMatchFunc func(body []byte) bool

// Request represents a mocked http request.
type Request struct {
	httpReq         http.Request
	respInfo        *Response
	expectedBody    []byte
	expectedHeaders http.Header
	matchBodyFunc   RequestBodyMatchFunc
	timesExpected   int
	hits            int
	onError         func(err error)
}

// newRequest returns a new mocked http request.
func newRequest() *Request {
	return &Request{
		httpReq: http.Request{
			URL: &url.URL{},
		},
		timesExpected: 1,
	}
}

// method sets the method and path of the request.
func (r *Request) method(method, path string) *Request {
	r.httpReq.Method = strings.ToUpper(method)

	u, err := url.Parse(path)
	if err != nil {
		r.onError(err)
		return nil
	}

	r.httpReq.URL.Path = u.Path
	r.httpReq.URL.RawQuery = u.RawQuery
	return r
}

// Times sets the expected number of times the request will be called.
// If not set, it defaults to 1.
func (r *Request) Times(times int) *Request {
	r.timesExpected = times
	return r
}

// MatchParam sets an expected query parameter that will be matched.
func (r *Request) MatchParam(key, value string) *Request {
	q := r.httpReq.URL.Query()
	q.Add(key, value)
	r.httpReq.URL.RawQuery = q.Encode()
	return r
}

// MatchHeader sets an expected header that will be matched.
func (r *Request) MatchHeader(key, value string) *Request {
	if r.expectedHeaders == nil {
		r.expectedHeaders = make(http.Header)
	}
	r.expectedHeaders.Add(key, value)
	return r
}

// MatchBody sets the expected value for the request body.
func (r *Request) MatchBody(body []byte) *Request {
	r.expectedBody = body
	return r
}

// MatchBodyFunc sets the expected value for the request body.
func (r *Request) MatchBodyFunc(fn RequestBodyMatchFunc) *Request {
	r.matchBodyFunc = fn
	return r
}

// Reply sets the expected status code and returns a Response object.
func (r *Request) Reply(status int) *Response {
	r.respInfo = &Response{reqInfo: r}
	r.respInfo.statusCode = status
	r.respInfo.header = make(http.Header)
	return r.respInfo
}

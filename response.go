package rexpect

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"time"
)

// ResponseModifierFunc represents a function to modify the response before it is returned.
type ResponseModifierFunc func(*http.Response)

// Response represents the response to be returned for a given request.
type Response struct {
	reqInfo    *Request
	statusCode int
	header     http.Header
	body       []byte
	delay      time.Duration
	modifiers  []ResponseModifierFunc
}

// Modify adds a function to modify the response before it is returned.
// The function can be used to modify the status code, headers or body of the
// response, or execute any code.
func (r *Response) Modify(fn ResponseModifierFunc) *Response {
	if fn != nil {
		r.modifiers = append(r.modifiers, fn)
	}
	return r
}

// Delay sets the delay for the response.
func (r *Response) Delay(delay time.Duration) *Response {
	r.delay = delay
	return r
}

// Body sets the response body.
func (r *Response) Body(body io.Reader) *Response {
	content, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}
	r.body = content
	return r
}

// BodyString sets the response body using a string.
func (r *Response) BodyString(body string) *Response {
	content, err := readAndDecode(body, "")
	if err != nil {
		panic(err)
	}
	r.body = content
	return r
}

// JSON sets the response body to a JSON representation of data
// and the Content-Type header to "application/json".
func (r *Response) JSON(data any) *Response {
	body, err := readAndDecode(data, "json")
	if err != nil {
		panic(err)
	}
	r.body = body
	r.header.Set("Content-Type", "application/json")
	return r
}

// XML sets the response body to an XML representation of data
// and the Content-Type header to "application/xml".
func (r *Response) XML(data any) *Response {
	body, err := readAndDecode(data, "xml")
	if err != nil {
		panic(err)
	}
	r.body = body
	r.header.Set("Content-Type", "application/xml")
	return r
}

// AddHeader adds a header key-value pair to the response.
func (r *Response) AddHeader(key, value string) *Response {
	r.header.Add(key, value)
	return r
}

// readAndDecode converts data to a byte array using the given
// encoding ("json" or "xml").
func readAndDecode(data any, kind string) ([]byte, error) {
	buf := &bytes.Buffer{}

	switch data := data.(type) {
	case string:
		buf.WriteString(data)
	case []byte:
		buf.Write(data)
	default:
		var err error
		if kind == "xml" {
			err = xml.NewEncoder(buf).Encode(data)
		} else {
			err = json.NewEncoder(buf).Encode(data)
		}
		if err != nil {
			return nil, err
		}
	}

	return io.ReadAll(buf)
}

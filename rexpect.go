package rexpect

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type HTTPHandler func(w http.ResponseWriter, r *http.Request)

// Rexpect is a simple HTTP mocking tool to mock out external HTTP requests
// in tests.
type Rexpect struct {
	server           *httptest.Server
	expectedCalls    []*Request
	requestsReceived []*http.Request
	onError          HTTPHandler
	t                *testing.T
}

// New creates a new Rexpect instance.
// You should call Stop() to stop the HTTP server when you are done with it, or
// use NewWithTest() to automatically stop it when the test finishes.
func New() *Rexpect {
	return NewWithTest(nil)
}

// NewWithTest creates a new Rexpect instance with the given testing.T.
// The HTTP server will be automatically stopped when the test finishes.
func NewWithTest(t *testing.T) *Rexpect {
	s := &Rexpect{
		expectedCalls:    nil,
		requestsReceived: nil,
		t:                t,
	}

	if s.t != nil {
		t.Cleanup(func() {
			s.Stop()
		})
	}

	server := httptest.NewUnstartedServer(s.handler())
	server.Start()
	s.server = server
	return s
}

// ErrorHandler registers the error handler to use when the HTTP handler
// cannot match a request or an error occurs.
func (s *Rexpect) ErrorHandler(handler HTTPHandler) *Rexpect {
	s.onError = handler
	return s
}

// Stop stops the HTTP server.
func (s *Rexpect) Stop() {
	s.server.Close()
}

// HostURL returns the URL of the HTTP server.
func (s *Rexpect) HostURL() string {
	return s.server.URL
}

// Expect returns a new Expect instance to register expected requests.
func (s *Rexpect) Expect() *Expect {
	return &Expect{rx: s}
}

// newRequest registers a new request.
func (s *Rexpect) newRequest() *Request {
	req := newRequest()
	req.onError = func(err error) {
		panic(err)
	}
	s.expectedCalls = append(s.expectedCalls, req)
	return req
}

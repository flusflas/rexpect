package rexpect

import "net/http"

type Expect struct {
	rx *Rexpect
}

// Method registers a new request with the given method and path.
func (e *Expect) Method(method, path string) *Request {
	req := e.rx.newRequest()
	return req.method(method, path)
}

// Post registers a new POST request.
func (e *Expect) Post(path string) *Request {
	return e.Method(http.MethodPost, path)
}

// Get registers a new GET request.
func (e *Expect) Get(path string) *Request {
	return e.Method(http.MethodGet, path)
}

// Put registers a new PUT request.
func (e *Expect) Put(path string) *Request {
	return e.Method(http.MethodPut, path)
}

// Patch registers a new PATCH request.
func (e *Expect) Patch(path string) *Request {
	return e.Method(http.MethodPatch, path)
}

// Delete registers a new DELETE request.
func (e *Expect) Delete(path string) *Request {
	return e.Method(http.MethodDelete, path)
}

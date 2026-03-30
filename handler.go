package rexpect

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

// handler returns an http.HandlerFunc that will match requests to the
// expected calls and handle them.
func (s *Rexpect) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqMatch *Request

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			s.fail(w, r, fmt.Errorf("cannot read body (%s %s)", r.Method, r.URL.String()))
		}

		// Loop through the expected calls to find a match
		for _, req := range s.expectedCalls {
			if req.hits >= req.timesExpected {
				continue
			}

			if req.httpReq.Method != r.Method ||
				r.URL.Path != req.httpReq.URL.Path {
				continue
			}

			if req.expectedBody != nil && string(req.expectedBody) != string(reqBody) {
				continue
			}

			if req.matchBodyFunc != nil {
				if !req.matchBodyFunc(reqBody) {
					continue
				}
			}

			if len(req.expectedHeaders) > 0 {
				headerMatch := true
				for key, expValues := range req.expectedHeaders {
					values, ok := r.Header[key]
					if !ok {
						headerMatch = false
						break
					}

					if slices.Compare(expValues, values) != 0 {
						headerMatch = false
						break
					}
				}

				if !headerMatch {
					continue
				}
			}

			queryMatch := true
			for expKey, expValues := range req.httpReq.URL.Query() {
				values, ok := r.URL.Query()[expKey]
				if !ok {
					queryMatch = false
					break
				}

				if slices.Compare(expValues, values) != 0 {
					queryMatch = false
					break
				}
			}

			if !queryMatch {
				continue
			}

			reqMatch = req
			reqMatch.hits++
			break
		}

		if reqMatch == nil {
			s.fail(w, r, fmt.Errorf("unexpected call (%s %s)", r.Method, r.URL.String()))
			return
		}

		if reqMatch.respInfo.delay > 0 {
			timer := time.NewTimer(reqMatch.respInfo.delay)
			select {
			case <-timer.C:
			case <-r.Context().Done():
				if !timer.Stop() {
					<-timer.C
				}
			}
		}

		// Creates a http.Response instance to pass to the filters
		resp := &http.Response{
			StatusCode: reqMatch.respInfo.statusCode,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(reqMatch.respInfo.body)),
			Request:    r,
		}

		for key, value := range reqMatch.respInfo.header {
			for _, v := range value {
				resp.Header.Add(key, v)
			}
		}

		for _, modifierFunc := range reqMatch.respInfo.modifiers {
			modifierFunc(resp)
		}

		for key, value := range resp.Header {
			for _, v := range value {
				w.Header().Add(key, v)
			}
		}

		w.WriteHeader(resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.fail(w, r, nil)
		}

		_, err = w.Write(body)
		if err != nil {
			s.fail(w, r, nil)
		}
	}
}

// fail fails the request handling.
func (s *Rexpect) fail(w http.ResponseWriter, r *http.Request, v any) {
	if v == nil {
		v = fmt.Errorf("error reading response body (%s %s)", r.Method, r.URL.String())
	}

	if s.onError != nil {
		s.onError(w, r)
	} else if s.t != nil {
		s.t.Errorf("%v", v)
	} else {
		s.abort(w, r)
	}
}

// abort closes the connection.
func (s *Rexpect) abort(w http.ResponseWriter, r *http.Request) {
	hj, _ := w.(http.Hijacker)
	conn, _, _ := hj.Hijack()
	if err := conn.Close(); err != nil {
		s.fail(w, r, "cannot close connection")
	}
}

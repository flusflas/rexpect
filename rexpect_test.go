package rexpect_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/flusflas/rexpect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRexpect(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Get("/test/json").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	m.Expect().
		Get("/test/xml").
		Reply(201).
		XML([]string{"foo", "bar"})

	m.Expect().
		Get("/test/string").
		MatchParam("some_param", "123").
		Reply(200).
		AddHeader("Content-Type", "text/plain").
		BodyString("Lorem ipsum dolor")

	res, err := http.Get(m.HostURL() + "/test/json")
	assert.NoError(t, err)
	assertResponse(t, res, 200, `{"foo":"bar"}`)
	assert.Equal(t, res.Header.Get("Content-Type"), "application/json")

	res, err = http.Get(m.HostURL() + "/test/xml")
	assert.NoError(t, err)
	assertResponse(t, res, 201, "<string>foo</string><string>bar</string>")
	assert.Equal(t, res.Header.Get("Content-Type"), "application/xml")

	res, err = http.Get(m.HostURL() + "/test/string?some_param=123")
	assert.NoError(t, err)
	assertResponse(t, res, 200, "Lorem ipsum dolor")
	assert.Equal(t, res.Header.Get("Content-Type"), "text/plain")

	// The request has already been made
	res, err = http.Get(m.HostURL() + "/test/json")
	assert.Error(t, err)
	assert.Nil(t, res)

	// TODO: Verify that we don't have pending mocks
}

func TestMatchHeader(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Get("/test/json").
		MatchHeader("X-Test", "test").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	req, err := http.NewRequest(http.MethodGet, m.HostURL()+"/test/json", nil)
	require.NoError(t, err)
	req.Header.Add("X-Test", "test")
	req.Header.Add("X-Foo", "bar")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assertResponse(t, res, 200, `{"foo":"bar"}`)

	req, err = http.NewRequest(http.MethodGet, m.HostURL()+"/test/json", nil)
	require.NoError(t, err)

	res, err = http.DefaultClient.Do(req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestNoMatch(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Get("/test/json").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Get(m.HostURL() + "/test/foo")
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestMatchBody(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Post("/test/echo").
		MatchBody([]byte(`{"foo": "bar"}`)).
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Post(m.HostURL()+"/test/echo", "aplication/json", io.NopCloser(strings.NewReader(`{"foo": "foo"}`)))
	assert.Error(t, err)
	assert.Nil(t, res)

	res, err = http.Post(m.HostURL()+"/test/echo", "aplication/json", io.NopCloser(strings.NewReader(`{"foo": "bar"}`)))
	assert.NoError(t, err)
	assertResponse(t, res, 200, `{"foo":"bar"}`)
}

func TestMatchBodyFunc(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Post("/test/echo").
		MatchBodyFunc(func(body []byte) bool {
			return string(body) == `{"foo": "bar"}`
		}).
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Post(m.HostURL()+"/test/echo", "aplication/json", io.NopCloser(strings.NewReader(`{"foo": "foo"}`)))
	assert.Error(t, err)
	assert.Nil(t, res)

	res, err = http.Post(m.HostURL()+"/test/echo", "aplication/json", io.NopCloser(strings.NewReader(`{"foo": "bar"}`)))
	assert.NoError(t, err)
	assertResponse(t, res, 200, `{"foo":"bar"}`)
}

func TestTimes(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Get("/test/json").
		Times(2).
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Get(m.HostURL() + "/test/json")
	assert.NoError(t, err)
	assertResponse(t, res, 200, `{"foo":"bar"}`)

	res, err = http.Get(m.HostURL() + "/test/json")
	assert.NoError(t, err)
	assertResponse(t, res, 200, `{"foo":"bar"}`)

	res, err = http.Get(m.HostURL() + "/test/json")
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestModify(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Get("/test/json").
		Times(2).
		Reply(200).
		Modify(func(response *http.Response) {
			response.Header.Add("X-Test", "test")
		}).
		Modify(func(response *http.Response) {
			response.Body = io.NopCloser(strings.NewReader(`{"foo":"changed!"}`))
		}).
		JSON(map[string]string{"foo": "bar"})

	res, err := http.Get(m.HostURL() + "/test/json")
	assert.NoError(t, err)
	assertResponse(t, res, 200, `{"foo":"changed!"}`)
	assert.Equal(t, res.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, res.Header.Get("X-Test"), "test")
}

func TestErrorHandler(t *testing.T) {
	m := rexpect.New()

	m.Expect().
		Get("/test/json").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	m.ErrorHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, err := w.Write([]byte("error"))
		assert.NoError(t, err)
	})

	res, err := http.Get(m.HostURL() + "/test/foo")
	assert.NoError(t, err)
	assertResponse(t, res, 500, "error")
}

func assertResponse(t *testing.T, res *http.Response, expectedStatusCode int, expectedBody string) {
	assert.Equal(t, res.StatusCode, expectedStatusCode)
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(string(body)), expectedBody)
}

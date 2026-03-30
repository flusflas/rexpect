<h3 align="center"><b>rexpect</b></h3>
<p align="center">A tiny HTTP mocking helper for Go testing</p>

<p align="center">
  <img src="https://img.shields.io/badge/go%20version-%3E=1.21-F37F40.svg" alt="Go Version">
  <a href="https://github.com/flusflas/rexpect/blob/master/LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="MIT License"></a>
  <a href="https://pkg.go.dev/github.com/flusflas/rexpect"><img src="https://pkg.go.dev/badge/github.com/flusflas/rexpect" alt="Go Documentation"></a>
</p>

## 🧐 What is _rexpect_?

**rexpect** is a tiny HTTP expectation/mocking helper for Go tests.

It runs an `httptest.Server` and lets you declare which HTTP requests you expect
your code to make (method, path, query params, headers, body), and what responses
to return.

Since it runs a real HTTP server, you can use it to test any code that makes
HTTP requests (as long as you can configure the base URL to point to the rexpect
server) without needing to change the code under test to inject a custom HTTP
client or transport.

## Install

```bash
go get github.com/flusflas/rexpect
```

## Quick start

```go
package mypkg_test

import (
	"net/http"
	"testing"

	"github.com/flusflas/rexpect"
)

func TestClient(t *testing.T) {
	r := rexpect.NewWithTest(t)

	r.Expect().
		Get("/v1/user").
		MatchParam("id", "123").
		MatchHeader("Authorization", "Bearer token").
		Reply(200).
		JSON(map[string]any{"id": 123, "name": "Ada"})

	res, err := http.Get(r.HostURL() + "/v1/user?id=123")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
}
```

## How it works

- You create a `*rexpect.Rexpect` server (`New()` or `NewWithTest(t)`).
- You register expected calls with `Expect()`.
- Each expected call can match on:
  - path (required)
  - query params (optional)
  - headers (optional)
  - body (optional)
- Then you define the response via `Reply(status)` and response helpers.

When an incoming request does not match any pending expected call (or the expected call was already used up), the server will:

- call your custom error handler if you registered one, OR
- `t.Errorf(...)` if you used `NewWithTest(t)`, OR
- abort/close the connection (making the client request fail).

## Examples

### Match JSON request body

```go
r := rexpect.NewWithTest(t)

r.Expect().
	Post("/echo").
	MatchBody([]byte(`{"foo": "bar"}`)).
	Reply(200).
	JSON(map[string]string{"ok": "true"})
```

### Custom body matcher (e.g. contains a specific text)

```go
r.Expect().
	Post("/echo").
	MatchBodyFunc(func(b []byte) bool {
		// your own comparison logic
		return strings.Contains(string(b), "hello")
	}).
	Reply(200).
	BodyString("ok")
```

### Same request expected multiple times

```go
r.Expect().
	Get("/health").
	Times(3).
	Reply(200).
	BodyString("ok")
```

### Delay responses

```go
r.Expect().
	Get("/slow").
	Reply(200).
	Delay(200 * time.Millisecond).
	BodyString("ok")
```

## Notes / limitations

- Rexpect is intentionally lightweight. It does not (currently) have a built-in "verify all expected calls were made" step.
- Header matching is strict for each key you declare: the request must include that key with the exact value set(s) you added.
- Query param matching is strict for any params you declared.

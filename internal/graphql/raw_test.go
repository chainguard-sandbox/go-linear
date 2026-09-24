package graphql

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yamashou/gqlgenc/clientv2"
)

func TestParseRawResponse_Data(t *testing.T) {
	var out json.RawMessage
	err := parseRawResponse([]byte(`{"data":{"viewer":{"id":"u1"}}}`), 200, &out)
	if err != nil {
		t.Fatalf("parseRawResponse() error = %v", err)
	}
	if string(out) != `{"viewer":{"id":"u1"}}` {
		t.Errorf("out = %s", out)
	}
}

func TestParseRawResponse_GraphQLErrors(t *testing.T) {
	var out json.RawMessage
	err := parseRawResponse([]byte(`{"errors":[{"message":"boom","path":["viewer"]}]}`), 200, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	var er *clientv2.ErrorResponse
	if !errors.As(err, &er) {
		t.Fatalf("error is not *clientv2.ErrorResponse: %T", err)
	}
	if er.NetworkError != nil {
		t.Errorf("NetworkError should be nil for a 200 GraphQL error, got %+v", er.NetworkError)
	}
	if er.GqlErrors == nil || len(*er.GqlErrors) != 1 {
		t.Fatalf("GqlErrors = %+v, want 1", er.GqlErrors)
	}
}

func TestParseRawResponse_NonJSONBodyNonOK(t *testing.T) {
	// Non-JSON body with a non-2xx status → NetworkError carrying the body.
	err := parseRawResponse([]byte(`<html>502 Bad Gateway</html>`), 502, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var er *clientv2.ErrorResponse
	if !errors.As(err, &er) {
		t.Fatalf("error is not *clientv2.ErrorResponse: %T (%v)", err, err)
	}
	if er.NetworkError == nil || er.NetworkError.Code != 502 {
		t.Fatalf("NetworkError = %+v, want code 502", er.NetworkError)
	}
	if !strings.Contains(er.NetworkError.Message, "Bad Gateway") {
		t.Errorf("NetworkError.Message = %q, want it to include the body", er.NetworkError.Message)
	}
}

func TestParseRawResponse_NonJSONBodyOK(t *testing.T) {
	// Non-JSON body with a 2xx status → a decode error (not a network error).
	err := parseRawResponse([]byte(`not json`), 200, nil)
	if err == nil {
		t.Fatal("expected decode error")
	}
	var er *clientv2.ErrorResponse
	if errors.As(err, &er) {
		t.Fatalf("expected a plain decode error, got *clientv2.ErrorResponse: %v", err)
	}
}

func TestParseRawResponse_MalformedErrors(t *testing.T) {
	// "errors" present but not an array on a 2xx → parse error.
	err := parseRawResponse([]byte(`{"errors":{"message":"oops"}}`), 200, nil)
	if err == nil {
		t.Fatal("expected error for malformed errors array")
	}
}

func TestParseRawResponse_MalformedErrorsNonOKPreservesNetworkError(t *testing.T) {
	// A non-2xx response whose "errors" is a malformed object (not the spec's
	// array) must still surface the NetworkError, not drop it (S-1).
	err := parseRawResponse([]byte(`{"errors":{"message":"forbidden"}}`), 403, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var er *clientv2.ErrorResponse
	if !errors.As(err, &er) {
		t.Fatalf("want *clientv2.ErrorResponse (NetworkError preserved), got %T (%v)", err, err)
	}
	if er.NetworkError == nil || er.NetworkError.Code != 403 {
		t.Fatalf("NetworkError = %+v, want code 403", er.NetworkError)
	}
}

func TestRaw_MaxInt64LimitDoesNotOverflow(t *testing.T) {
	// WithMaxResponseBytes(math.MaxInt64) must not overflow limit+1 into a
	// negative LimitReader bound that reads zero bytes (S-4).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"x":"ok"}}`))
	}))
	defer srv.Close()

	c := &Client{Client: clientv2.NewClient(http.DefaultClient, srv.URL, nil)}
	var out json.RawMessage
	if err := c.Raw(context.Background(), "", `query { x }`, nil, &out, math.MaxInt64); err != nil {
		t.Fatalf("Raw() error = %v", err)
	}
	if string(out) != `{"x":"ok"}` {
		t.Errorf("out = %s", out)
	}
}

func TestRaw_OversizeResponseErrorsClearly(t *testing.T) {
	// A response larger than the (per-call) limit must fail with a clear size
	// error, not a confusing truncated-JSON decode error, and must not echo the
	// body. Uses a small explicit limit to exercise the override cheaply.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		big := bytes.Repeat([]byte("a"), 4096)
		_, _ = w.Write(append([]byte(`{"data":{"x":"`), big...))
	}))
	defer srv.Close()

	c := &Client{Client: clientv2.NewClient(http.DefaultClient, srv.URL, nil)}
	var out json.RawMessage
	err := c.Raw(context.Background(), "", `query { x }`, nil, &out, 1024) // 1KB cap
	if err == nil {
		t.Fatal("Raw() expected an oversize error, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("error = %q, want it to mention the size limit", err.Error())
	}
	if strings.Contains(err.Error(), "aaaa") {
		t.Errorf("error should not echo the response body")
	}
}

func TestRaw_UnderLimitWithExplicitCap(t *testing.T) {
	// A response under the explicit cap decodes normally.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"x":"ok"}}`))
	}))
	defer srv.Close()

	c := &Client{Client: clientv2.NewClient(http.DefaultClient, srv.URL, nil)}
	var out json.RawMessage
	if err := c.Raw(context.Background(), "", `query { x }`, nil, &out, 1024); err != nil {
		t.Fatalf("Raw() error = %v", err)
	}
	if string(out) != `{"x":"ok"}` {
		t.Errorf("out = %s", out)
	}
}

func TestRaw_GzipResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, _ = gz.Write([]byte(`{"data":{"viewer":{"id":"gz"}}}`))
		_ = gz.Close()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(buf.Bytes())
	}))
	defer srv.Close()

	// Disable transparent decompression so Raw's manual gzip branch runs.
	httpClient := &http.Client{Transport: &http.Transport{DisableCompression: true}}
	c := &Client{Client: clientv2.NewClient(httpClient, srv.URL, nil)}
	var out json.RawMessage
	if err := c.Raw(context.Background(), "", `query { viewer { id } }`, nil, &out, 0); err != nil {
		t.Fatalf("Raw() error = %v", err)
	}
	if string(out) != `{"viewer":{"id":"gz"}}` {
		t.Errorf("out = %s", out)
	}
}

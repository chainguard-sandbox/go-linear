// Package graphql contains the gqlgenc-generated Linear GraphQL client
// (models.go, client.go — do not edit) alongside the hand-written raw escape
// hatch in this file.
package graphql

// This file is hand-written (NOT codegen). The generated client.go carries a
// "DO NOT EDIT" header and defines the LinearGraphQLClient interface, so the raw
// escape hatch lives here on the concrete *Client instead.

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// defaultMaxRawResponseSize caps how many bytes of a response body Raw will read
// when the caller passes no explicit limit, so a hostile or misbehaving server
// cannot exhaust memory. The transport's MaxBodySize caps request bodies only,
// so the response read is bounded here.
const defaultMaxRawResponseSize = 10 * 1024 * 1024 // 10MB

// Raw executes an arbitrary GraphQL document through the underlying
// clientv2.Client, inheriting the exact same interceptor and transport stack as
// every generated operation (auth header injection, 401 refresh-and-retry,
// retries/backoff, circuit breaker, rate limiting, and metrics). The response
// body is read under a size limit: maxResponseBytes when positive, else
// defaultMaxRawResponseSize.
//
// out receives the response "data" object, decoded with the standard library
// json package. Unlike the generated operations, Raw deliberately does NOT route
// decoding through gqlgenc's graphqljson decoder: that decoder is struct-oriented
// and cannot unmarshal into an arbitrary top-level target, whereas a raw
// passthrough must accept *json.RawMessage, *map[string]any, or any struct.
//
// GraphQL errors are returned as *clientv2.ErrorResponse (with GqlErrors set) and
// network/HTTP errors as *clientv2.ErrorResponse (with NetworkError set), matching
// what clientv2.Post returns, so callers can classify them uniformly.
func (c *Client) Raw(ctx context.Context, opName, query string, vars map[string]any, out any, maxResponseBytes int64) error {
	limit := maxResponseBytes
	if limit <= 0 {
		limit = defaultMaxRawResponseSize
	}
	if limit == math.MaxInt64 {
		limit = math.MaxInt64 - 1 // avoid overflow on the limit+1 read below
	}

	reqPayload := &clientv2.Request{
		Query:         query,
		Variables:     vars,
		OperationName: opName,
	}

	body, err := clientv2.MarshalJSON(ctx, reqPayload)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Client.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	// Terminal handler: performs the HTTP call through the configured http.Client
	// (whose Transport carries retries, circuit breaker, rate limiting and
	// metrics) and parses the GraphQL envelope ourselves.
	do := func(_ context.Context, r *http.Request, _ *clientv2.GQLRequestInfo, res any) error {
		resp, doErr := c.Client.Client.Do(r)
		if doErr != nil {
			return fmt.Errorf("request failed: %w", doErr)
		}
		defer resp.Body.Close()

		reader := io.Reader(resp.Body)
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gz, gzErr := gzip.NewReader(resp.Body)
			if gzErr != nil {
				return fmt.Errorf("gzip decode failed: %w", gzErr)
			}
			defer gz.Close()
			reader = gz
		}

		// Bound the response read so a hostile/misbehaving server (or a custom
		// LINEAR_BASE_URL) cannot exhaust memory; also bounds the decompressed
		// size of a gzip bomb. Read one byte past the limit to distinguish a body
		// that is exactly at the cap from one that overflows it, and fail with a
		// clear error rather than silently truncating into a JSON decode failure.
		respBody, readErr := io.ReadAll(io.LimitReader(reader, limit+1))
		if readErr != nil {
			return fmt.Errorf("failed to read response body: %w", readErr)
		}
		if int64(len(respBody)) > limit {
			return fmt.Errorf("graphql response exceeds %d-byte limit", limit)
		}

		return parseRawResponse(respBody, resp.StatusCode, res)
	}

	// Run through the client's interceptor chain (auth header + 401 refresh) with
	// our terminal handler, mirroring how clientv2.Post drives requests.
	gqlInfo := clientv2.NewGQLRequestInfo(reqPayload)
	return c.Client.RequestInterceptor(ctx, req, gqlInfo, out, do)
}

// parseRawResponse decodes a GraphQL HTTP response, returning a
// *clientv2.ErrorResponse for network or GraphQL errors (so downstream
// classification is identical to clientv2.Post) and otherwise unmarshaling the
// "data" object into out.
func parseRawResponse(body []byte, httpCode int, out any) error {
	errResponse := &clientv2.ErrorResponse{}
	isOKCode := 200 <= httpCode && httpCode <= 299
	if !isOKCode {
		errResponse.NetworkError = &clientv2.HTTPError{
			Code:    httpCode,
			Message: fmt.Sprintf("Response body %s", truncateForError(body)),
		}
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors json.RawMessage `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		// Non-JSON body (e.g. a gateway error page). If the HTTP status already
		// flagged a network error, surface that; otherwise report the decode error.
		if errResponse.HasErrors() {
			return errResponse
		}
		return fmt.Errorf("failed to decode response %s: %w", truncateForError(body), err)
	}

	if len(envelope.Errors) > 0 {
		var list gqlerror.List
		if err := json.Unmarshal(envelope.Errors, &list); err != nil {
			// A malformed "errors" value (e.g. an object rather than the spec's
			// array, from an intermediary). Don't abandon a non-2xx NetworkError:
			// fall through to HasErrors() so status classification is preserved.
			if errResponse.HasErrors() {
				return errResponse
			}
			return fmt.Errorf("failed to parse graphql errors %s: %w", truncateForError(body), err)
		}
		errResponse.GqlErrors = &list
	}

	if errResponse.HasErrors() {
		return errResponse
	}

	if out != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("failed to decode data into out: %w", err)
		}
	}

	return nil
}

// truncateForError bounds how much of a response body is embedded in an error
// message, so a large or sensitive body is not echoed wholesale to callers/logs.
func truncateForError(body []byte) string {
	const maxLen = 512
	if len(body) > maxLen {
		return string(body[:maxLen]) + "…(truncated)"
	}
	return string(body)
}

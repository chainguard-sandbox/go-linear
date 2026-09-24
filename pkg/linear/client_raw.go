package linear

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// rawExecutor is the escape-hatch capability satisfied by *intgraphql.Client via
// its hand-written Raw method (internal/graphql/raw.go).
type rawExecutor interface {
	Raw(ctx context.Context, opName, query string, vars map[string]any, out any, maxResponseBytes int64) error
}

// ExecuteOption configures a single Execute call.
type ExecuteOption func(*executeConfig)

type executeConfig struct {
	operationName    string
	maxResponseBytes int64
}

// WithOperationName selects which operation to run when the document defines
// more than one. It maps to the request's "operationName" field. Leave it unset
// (the default) for a document with a single (or anonymous) operation; the server
// then runs that sole operation.
func WithOperationName(name string) ExecuteOption {
	return func(c *executeConfig) { c.operationName = name }
}

// WithMaxResponseBytes caps how many bytes of the response body this call will
// read before failing with a clear size error, guarding against a hostile or
// misbehaving server exhausting memory. A non-positive value (the default) uses
// the built-in 10MB limit. Raise it for a query you expect to return a large
// un-modeled result set.
func WithMaxResponseBytes(n int64) ExecuteOption {
	return func(c *executeConfig) { c.maxResponseBytes = n }
}

// GraphQLResponseError is a typed error carrying one or more GraphQL errors
// returned in a response's "errors" array. It is distinct from network/HTTP
// errors (*AuthenticationError, *ForbiddenError, *LinearError of type
// NetworkError, etc.), letting callers tell a well-formed GraphQL-level failure
// apart from a transport failure.
//
// Example:
//
//	var raw json.RawMessage
//	err := client.Execute(ctx, query, vars, &raw)
//	var gqlErr *linear.GraphQLResponseError
//	if errors.As(err, &gqlErr) {
//	    for _, e := range gqlErr.Errors {
//	        log.Printf("%s (path=%v code=%v)", e.Message, e.Path, e.Extensions["code"])
//	    }
//	}
type GraphQLResponseError struct {
	Errors []GraphQLError
}

// Error implements the error interface, rendering each GraphQL error's message
// and (when present) its path.
func (e *GraphQLResponseError) Error() string {
	msg := "unspecified graphql error"
	if len(e.Errors) > 0 {
		parts := make([]string, 0, len(e.Errors))
		for _, ge := range e.Errors {
			if len(ge.Path) > 0 {
				parts = append(parts, fmt.Sprintf("%s (path: %v)", ge.Message, ge.Path))
			} else {
				parts = append(parts, ge.Message)
			}
		}
		msg = strings.Join(parts, "; ")
	}
	return fmt.Sprintf("linear: %s: %s", ErrorTypeGraphQLError, msg)
}

// Execute runs an arbitrary GraphQL document through the fully-configured client,
// inheriting authentication, retries, rate limiting, metrics and credential
// rotation exactly like every typed SDK method. It is a supported escape hatch
// for requesting un-modeled fields, filters beyond the builders, or brand-new API
// surface.
//
// out is unmarshaled from the response "data" object using the standard library
// json package, so it accepts *json.RawMessage, *map[string]any, or a typed
// struct. Pass nil to discard the data.
//
// A 2xx response carrying an "errors" array surfaces as *GraphQLResponseError
// (preserving message, path and extensions). A non-2xx / transport failure —
// authentication (401), authorization (403), rate-limit (429), etc. — is
// classified by status code into the usual typed errors. To run a specific
// operation from a multi-operation document, pass WithOperationName; to change
// the response size limit, pass WithMaxResponseBytes.
//
// This is a raw passthrough: there is NO name→ID resolution, NO field defaults,
// NO --fields pruning, and NO complexity reduction. The response shape is exactly
// what Linear returns and is not stable across API changes.
func (c *Client) Execute(ctx context.Context, query string, vars map[string]any, out any, opts ...ExecuteOption) error {
	exec, ok := c.gqlClient.(rawExecutor)
	if !ok {
		return &LinearError{
			Type:    ErrorTypeInternalError,
			Message: "raw GraphQL execution not supported by this client",
		}
	}

	var cfg executeConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := exec.Raw(ctx, cfg.operationName, query, vars, out, cfg.maxResponseBytes); err != nil {
		// Classify by structure first, not by string matching. A non-2xx response
		// carries a NetworkError classified by its status CODE (not by matching the
		// body text, which would mistake a 500 whose body says "FORBIDDEN" for a
		// 403). A 2xx response carrying only errors[] is a GraphQL-level failure
		// and keeps its structured detail even when a message contains such a token.
		if er, ok := errors.AsType[*clientv2.ErrorResponse](err); ok {
			if er.NetworkError != nil {
				code := er.NetworkError.Code
				// Auth/permission/rate-limit are security-relevant: classify by code.
				if code == 401 || code == 403 || code == 429 {
					return wrapNetworkStatus(code, err)
				}
				// Other non-2xx: prefer the structured GraphQL errors when present
				// (e.g. a 400/500 validation response), else a status-coded error.
				if er.GqlErrors != nil {
					return &GraphQLResponseError{Errors: convertGQLErrors(*er.GqlErrors)}
				}
				return wrapNetworkStatus(code, err)
			}
			if er.GqlErrors != nil {
				return &GraphQLResponseError{Errors: convertGQLErrors(*er.GqlErrors)}
			}
		}
		// Local errors from the raw executor (encode, read, response-size limit,
		// decode) are already clean and specific — preserve the message rather than
		// flattening it to a generic "graphql failed".
		return &LinearError{
			Type:    ErrorTypeGraphQLError,
			Message: err.Error(),
			wrapped: err,
		}
	}

	return nil
}

// convertGQLErrors maps gqlgenc/gqlparser errors into the SDK's GraphQLError type.
func convertGQLErrors(list gqlerror.List) []GraphQLError {
	out := make([]GraphQLError, 0, len(list))
	for _, e := range list {
		if e == nil {
			continue
		}
		var path []any
		if len(e.Path) > 0 {
			path = make([]any, 0, len(e.Path))
			for _, p := range e.Path {
				path = append(path, p)
			}
		}
		out = append(out, GraphQLError{
			Message:    e.Message,
			Path:       path,
			Extensions: e.Extensions,
		})
	}
	return out
}

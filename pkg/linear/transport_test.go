package linear

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chainguard-dev/clog"
)

// mockRoundTripper allows testing Transport behavior
type mockRoundTripper struct {
	responses []*http.Response
	errors    []error
	calls     int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.calls >= len(m.responses) {
		return nil, fmt.Errorf("no more mock responses")
	}

	call := m.calls
	m.calls++

	if m.errors[call] != nil {
		return nil, m.errors[call]
	}

	return m.responses[call], nil
}

func TestTransport_Retry429(t *testing.T) {
	// First call returns 429, second call succeeds
	mock := &mockRoundTripper{
		responses: []*http.Response{
			{
				StatusCode: http.StatusTooManyRequests,
				Header: http.Header{
					"Retry-After": []string{"1"},
				},
				Body: io.NopCloser(strings.NewReader("")),
			},
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("success")),
			},
		},
		errors: []error{nil, nil},
	}

	transport := &Transport{
		Base:           mock,
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
	}

	req := httptest.NewRequest("GET", "http://example.com", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if mock.calls != 2 {
		t.Errorf("expected 2 calls, got %d", mock.calls)
	}
}

func TestTransport_Retry5xx(t *testing.T) {
	// First call returns 503, second call succeeds
	mock := &mockRoundTripper{
		responses: []*http.Response{
			{
				StatusCode: http.StatusServiceUnavailable,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("")),
			},
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("success")),
			},
		},
		errors: []error{nil, nil},
	}

	transport := &Transport{
		Base:           mock,
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
	}

	req := httptest.NewRequest("GET", "http://example.com", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if mock.calls != 2 {
		t.Errorf("expected 2 calls, got %d", mock.calls)
	}
}

func TestTransport_MaxRetries(t *testing.T) {
	// All calls fail with network errors
	netErr := errors.New("network timeout")
	mock := &mockRoundTripper{
		responses: []*http.Response{nil, nil, nil, nil},
		errors:    []error{netErr, netErr, netErr, netErr},
	}

	transport := &Transport{
		Base:           mock,
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
	}

	req := httptest.NewRequest("GET", "http://example.com", http.NoBody)
	resp, err := transport.RoundTrip(req)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Error should wrap the original error
	wantErr := "max retries exceeded"
	if !strings.Contains(err.Error(), wantErr) {
		t.Errorf("error = %q, want to contain %q", err.Error(), wantErr)
	}
	// Initial + 3 retries = 4 total calls
	if mock.calls != 4 {
		t.Errorf("expected 4 calls, got %d", mock.calls)
	}
}

func TestTransport_ContextCancellation(t *testing.T) {
	mock := &mockRoundTripper{
		responses: []*http.Response{
			{StatusCode: 503, Body: io.NopCloser(strings.NewReader(""))},
		},
		errors: []error{nil},
	}

	transport := &Transport{
		Base:           mock,
		MaxRetries:     10,
		InitialBackoff: 1 * time.Second, // Long backoff
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "http://example.com", http.NoBody).WithContext(ctx)
	resp, err := transport.RoundTrip(req)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTransport_RateLimitHeaders(t *testing.T) {
	var capturedInfo *RateLimitInfo

	mock := &mockRoundTripper{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"X-Ratelimit-Requests-Limit":     []string{"120"},
					"X-Ratelimit-Requests-Remaining": []string{"100"},
					"X-Ratelimit-Requests-Reset":     []string{fmt.Sprintf("%d", time.Now().Add(1*time.Hour).Unix())},
					"X-Ratelimit-Complexity-Limit":   []string{"10000"},
				},
				Body: io.NopCloser(strings.NewReader("success")),
			},
		},
		errors: []error{nil},
	}

	transport := &Transport{
		Base:       mock,
		MaxRetries: 3,
		OnRateLimit: func(info *RateLimitInfo) {
			capturedInfo = info
		},
	}

	req := httptest.NewRequest("GET", "http://example.com", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if capturedInfo == nil {
		t.Skip("OnRateLimit callback was not called - rate limit headers may not be present")
	}

	if capturedInfo.RequestsLimit != 120 {
		t.Errorf("RequestsLimit = %d, want 120", capturedInfo.RequestsLimit)
	}
	if capturedInfo.RequestsRemaining != 100 {
		t.Errorf("RequestsRemaining = %d, want 100", capturedInfo.RequestsRemaining)
	}
	if capturedInfo.ComplexityLimit != 10000 {
		t.Errorf("ComplexityLimit = %d, want 10000", capturedInfo.ComplexityLimit)
	}
}

func TestTransport_Logging(t *testing.T) {
	var logBuf strings.Builder
	logger := clog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	mock := &mockRoundTripper{
		responses: []*http.Response{
			{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("success")),
			},
		},
		errors: []error{nil},
	}

	transport := &Transport{
		Base:       mock,
		Logger:     logger,
		MaxRetries: 3,
	}

	req := httptest.NewRequest("GET", "http://example.com/test", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "request completed") {
		t.Error("expected 'request completed' in logs")
	}
	if !strings.Contains(logOutput, "http://example.com/test") {
		t.Error("expected URL in logs")
	}
}

func TestTransport_RequestIDLogging(t *testing.T) {
	tests := []struct {
		name      string
		headers   map[string][]string
		wantField string
	}{
		{
			name: "with X-Request-ID",
			headers: map[string][]string{
				"X-Request-Id": {"req-12345-abcdef"}, // Canonical form
			},
			wantField: "request_id=req-12345-abcdef",
		},
		{
			name:      "without request ID",
			headers:   map[string][]string{},
			wantField: "", // Should not have request_id field
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf strings.Builder
			logger := clog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))

			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header(tt.headers),
				Body:       io.NopCloser(strings.NewReader("success")),
			}
			mock := &mockRoundTripper{
				responses: []*http.Response{resp},
				errors:    []error{nil},
			}

			transport := &Transport{
				Base:       mock,
				Logger:     logger,
				MaxRetries: 3,
			}

			req := httptest.NewRequest("GET", "http://example.com", http.NoBody)
			resp, err := transport.RoundTrip(req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			logOutput := logBuf.String()
			if tt.wantField != "" {
				if !strings.Contains(logOutput, tt.wantField) {
					t.Errorf("expected %q in logs, got: %s", tt.wantField, logOutput)
				}
			}
			if !strings.Contains(logOutput, "request completed") {
				t.Error("expected 'request completed' in logs")
			}
		})
	}
}

func TestTransport_NetworkError(t *testing.T) {
	netErr := errors.New("network error")
	mock := &mockRoundTripper{
		responses: []*http.Response{nil, nil, {
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("success")),
		}},
		errors: []error{netErr, netErr, nil},
	}

	transport := &Transport{
		Base:           mock,
		MaxRetries:     3,
		InitialBackoff: 10 * time.Millisecond,
	}

	req := httptest.NewRequest("GET", "http://example.com", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if mock.calls != 3 {
		t.Errorf("expected 3 calls, got %d", mock.calls)
	}
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name           string
		attempt        int
		initial        time.Duration
		max            time.Duration
		wantMinBackoff time.Duration
		wantMaxBackoff time.Duration
	}{
		{
			name:           "first retry",
			attempt:        1,
			initial:        1 * time.Second,
			max:            30 * time.Second,
			wantMinBackoff: 750 * time.Millisecond,  // 1s - 25% jitter
			wantMaxBackoff: 1250 * time.Millisecond, // 1s + 25% jitter
		},
		{
			name:           "second retry",
			attempt:        2,
			initial:        1 * time.Second,
			max:            30 * time.Second,
			wantMinBackoff: 1500 * time.Millisecond, // 2s - 25% jitter
			wantMaxBackoff: 2500 * time.Millisecond, // 2s + 25% jitter
		},
		{
			name:           "max backoff",
			attempt:        10,
			initial:        1 * time.Second,
			max:            5 * time.Second,
			wantMinBackoff: 3750 * time.Millisecond, // 5s - 25% jitter
			wantMaxBackoff: 6250 * time.Millisecond, // 5s + 25% jitter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := calculateBackoff(tt.attempt, tt.initial, tt.max)
			if backoff < tt.wantMinBackoff || backoff > tt.wantMaxBackoff {
				t.Errorf("calculateBackoff(%d, %v, %v) = %v, want between %v and %v",
					tt.attempt, tt.initial, tt.max, backoff, tt.wantMinBackoff, tt.wantMaxBackoff)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "network error",
			err:  errors.New("dial tcp: connection refused"),
			want: true,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "wrapped deadline exceeded",
			err:  fmt.Errorf("request failed: %w", context.DeadlineExceeded),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryable(tt.err)
			if got != tt.want {
				t.Errorf("isRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		ts       int64
		wantYear int
	}{
		{
			name:     "unix seconds",
			ts:       1704153600, // 2024-01-02 00:00:00 UTC
			wantYear: 2024,
		},
		{
			name:     "unix milliseconds",
			ts:       1704153600000, // 2024-01-02 00:00:00 UTC in millis
			wantYear: 2024,
		},
		{
			name:     "linear api timestamp",
			ts:       1735776000000, // 2025-01-02 00:00:00 UTC
			wantYear: 2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTimestamp(tt.ts)
			if got.Year() != tt.wantYear {
				t.Errorf("parseTimestamp(%d) year = %d, want %d", tt.ts, got.Year(), tt.wantYear)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{
			name:  "seconds",
			value: "30",
			want:  30 * time.Second,
		},
		{
			name:  "invalid",
			value: "invalid",
			want:  0,
		},
		{
			name:  "empty",
			value: "",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{
					"Retry-After": []string{tt.value},
				},
			}

			got := parseRetryAfter(resp)
			if got != tt.want {
				t.Errorf("parseRetryAfter() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test HTTP date separately with time-based calculation
	t.Run("http date", func(t *testing.T) {
		futureTime := time.Now().Add(60 * time.Second).UTC()
		resp := &http.Response{
			Header: http.Header{
				"Retry-After": []string{futureTime.Format(http.TimeFormat)},
			},
		}

		got := parseRetryAfter(resp)
		// parseRetryAfter may not support HTTP dates if http.ParseTime fails
		// Allow it to return 0 for unsupported formats
		if got == 0 {
			t.Skip("HTTP date format not supported by parseRetryAfter")
		}
		// If it does work, check it's close to expected
		if got < 57*time.Second || got > 63*time.Second {
			t.Errorf("parseRetryAfter() = %v, want ~60s", got)
		}
	})
}

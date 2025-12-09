# Operational Runbook

This runbook covers common incidents and their resolution procedures for go-linear in production.

## Quick Reference

| Symptom | Likely Cause | Runbook Section |
|---------|--------------|-----------------|
| Error rate spike | Linear API down or degraded | [Linear API Errors](#linear-api-returning-errors) |
| `rate_limit_remaining` < 10% | High request volume | [Rate Limits](#hitting-rate-limits) |
| p95 latency > 5s | Slow API or retries | [Slow Responses](#slow-api-responses) |
| "too many open files" | Connection pool exhaustion | [Connection Pool](#connection-pool-exhaustion) |

---

## Incidents

### Linear API Returning Errors

**Detection:**
- Error rate spike in `linear_errors_total` metric
- `request failed` errors in logs
- HTTP 5xx status codes in logs

**Investigation:**
1. Check error type in logs (authentication, rate limit, or API error)
2. Extract `request_id` from structured logs
3. Check Linear API status: https://linear.statuspage.io
4. Contact Linear support with request ID if persistent

**Logs to check:**
```json
{
  "level": "ERROR",
  "msg": "request failed",
  "status": 503,
  "request_id": "req-abc-123",
  "error": "server error (503)"
}
```

**Mitigation:**
- **Automatic:** Client retries 5xx errors automatically (up to 3 attempts)
- **Manual:** If prolonged (>5 min), implement circuit breaker or fallback
- **Escalation:** Contact Linear support with request_id for assistance

**Post-Incident:**
- Review retry metrics: `linear_retries_total{reason="server_error"}`
- Check if retries succeeded or all failed
- Adjust `MaxRetries` or `MaxRetryDuration` if needed

---

### Hitting Rate Limits

**Detection:**
- `linear_rate_limit_remaining` gauge < 10% of limit
- 429 status codes in logs
- `rate limited, waiting` log messages
- `linear_retries_total{reason="rate_limited"}` increasing

**Investigation:**
1. Check `RequestsRemaining` vs `RequestsLimit` in logs:
   ```json
   {
     "level": "DEBUG",
     "msg": "rate limit info",
     "requests_remaining": 10,
     "requests_limit": 120
   }
   ```
2. Identify high-volume operations in `linear_requests_total{operation}`
3. Check for burst traffic patterns (sudden spikes)
4. Review `ComplexityRemaining` for GraphQL query complexity limits

**Mitigation:**
1. **Immediate:** Client auto-retries with `Retry-After` header - wait for automatic recovery
2. **Short-term:** Reduce request frequency or batch operations
3. **Long-term:**
   - Implement client-side rate limiting
   - Request rate limit increase from Linear support
   - Optimize GraphQL queries to reduce complexity

**Prevention:**
```go
// Add rate limit callback for proactive throttling
linear.WithRateLimitCallback(func(info *linear.RateLimitInfo) {
    if info.RequestsRemaining < info.RequestsLimit/10 {
        // Throttle application requests
        time.Sleep(1 * time.Second)
    }
})
```

---

### Slow API Responses

**Detection:**
- `linear_request_duration_seconds` p95 > 5s
- `duration` field in logs > 5s
- User reports of timeouts

**Investigation:**
1. Check `duration` in structured logs
2. Look for retry attempts in logs (indicates failures → retries → latency)
3. Check Linear API status for degradation
4. Review `linear_retries_total` for spike in retries

**Example slow request log:**
```json
{
  "level": "INFO",
  "msg": "request completed",
  "status": 200,
  "duration": "8.5s",
  "request_id": "req-xyz-789"
}
```

**Mitigation:**
1. **Immediate:** Increase client timeout if needed:
   ```go
   linear.WithTimeout(60*time.Second)
   ```
2. **Short-term:** Reduce retry attempts to fail faster:
   ```go
   linear.WithRetry(1, 500*time.Millisecond, 5*time.Second)
   ```
3. **Long-term:**
   - Implement circuit breaker to stop retrying during outages
   - Add request prioritization (critical vs best-effort)
   - Cache frequently accessed data

**Post-Incident:**
- Contact Linear support with slow `request_id` values
- Review query complexity if using custom GraphQL queries

---

### Connection Pool Exhaustion

**Detection:**
- Errors: "too many open files" or "dial tcp: cannot assign requested address"
- File descriptor limit reached
- Application hanging on new requests

**Investigation:**
1. Check number of go-linear client instances created
2. Verify `client.Close()` is called on all clients
3. Check file descriptor limits: `ulimit -n`
4. Count open connections: `lsof -p <pid> | grep ESTABLISHED | wc -l`

**Mitigation:**
1. **Immediate:** Restart application to close leaked connections
2. **Code fix:** Ensure `defer client.Close()` after all `NewClient()` calls
3. **Configuration:**
   - Reduce `MaxIdleConnsPerHost` (default: 3):
     ```go
     transport := &http.Transport{
         MaxIdleConnsPerHost: 2,
     }
     linear.WithTransport(&linear.Transport{Base: transport})
     ```
   - Increase file descriptor limit: `ulimit -n 4096`
4. **Architecture:** Implement client pooling instead of per-request clients

**Prevention:**
```go
// Always use defer
client, err := linear.NewClient(apiKey)
if err != nil {
    return err
}
defer client.Close() // REQUIRED

// Or use with context for automatic cleanup
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

---

## Emergency Contacts

- **Linear Support:** #api channel on Linear Slack or support@linear.app
- **Status Page:** https://linear.statuspage.io
- **API Documentation:** https://developers.linear.app

## Escalation Criteria

**Page Oncall:** Error rate > 50% for 5+ minutes
**Engage Linear Support:** Persistent API errors with request_id for > 15 minutes
**Implement Circuit Breaker:** Error rate > 10% for > 30 minutes

# Monitoring Guide

This guide covers Prometheus metrics, alerts, and dashboards for go-linear.

## Required Metrics

### RED Metrics (Rate, Errors, Duration)

```promql
# Request Rate
linear_requests_total{operation, status_code}

# Error Rate
linear_errors_total{operation, error_type}

# Duration (Latency)
linear_request_duration_seconds{operation}
```

### Rate Limiting Metrics

```promql
# Remaining capacity (gauge)
linear_rate_limit_remaining{limit_type}  # limit_type: "requests" or "complexity"

# Total retries due to rate limiting
linear_retries_total{reason="rate_limited"}
```

### Retry Metrics

```promql
# Retry attempts by reason
linear_retries_total{reason}  # reason: "rate_limited", "server_error", "network_error"
```

---

## Prometheus Queries

### Request Rate (Requests per Second)

```promql
# Total request rate
rate(linear_requests_total[5m])

# Success rate
rate(linear_requests_total{status_code=~"2.."}[5m])

# Error rate
rate(linear_requests_total{status_code=~"[45].."}[5m])
```

### Error Rate Percentage

```promql
# Error percentage
(
  sum(rate(linear_errors_total[5m]))
  /
  sum(rate(linear_requests_total[5m]))
) * 100
```

### Latency Percentiles

```promql
# p50 latency
histogram_quantile(0.50, rate(linear_request_duration_seconds_bucket[5m]))

# p95 latency
histogram_quantile(0.95, rate(linear_request_duration_seconds_bucket[5m]))

# p99 latency
histogram_quantile(0.99, rate(linear_request_duration_seconds_bucket[5m]))
```

### Rate Limit Utilization

```promql
# Rate limit utilization percentage
(1 - (linear_rate_limit_remaining{limit_type="requests"} / 120)) * 100

# Complexity utilization percentage
(1 - (linear_rate_limit_remaining{limit_type="complexity"} / 10000)) * 100
```

### Retry Rate

```promql
# Retry percentage
(
  sum(rate(linear_retries_total[5m]))
  /
  sum(rate(linear_requests_total[5m]))
) * 100

# Retry breakdown by reason
sum by(reason) (rate(linear_retries_total[5m]))
```

---

## Alert Rules

### Critical Alerts (Page Oncall)

```yaml
groups:
  - name: linear_critical
    interval: 30s
    rules:
      - alert: LinearAPIDown
        expr: |
          (
            sum(rate(linear_errors_total[5m]))
            /
            sum(rate(linear_requests_total[5m]))
          ) > 0.5
        for: 5m
        labels:
          severity: P1
          component: linear_client
        annotations:
          summary: "Linear API error rate > 50%"
          description: "Error rate is {{ $value | humanizePercentage }}. Check Linear status page and logs for request_id."
          runbook: "docs/RUNBOOK.md#linear-api-returning-errors"

      - alert: LinearAllRequestsFailing
        expr: |
          sum(rate(linear_requests_total{status_code=~"2.."}[5m])) == 0
          and
          sum(rate(linear_requests_total[5m])) > 0
        for: 3m
        labels:
          severity: P1
          component: linear_client
        annotations:
          summary: "All Linear API requests failing"
          description: "Zero successful requests in last 3 minutes. Likely authentication or network issue."
```

### Warning Alerts (Slack/Email)

```yaml
  - name: linear_warnings
    interval: 1m
    rules:
      - alert: LinearRateLimitApproaching
        expr: |
          linear_rate_limit_remaining{limit_type="requests"} < 12
        for: 5m
        labels:
          severity: P3
          component: linear_client
        annotations:
          summary: "Linear API rate limit < 10%"
          description: "Only {{ $value }} requests remaining out of 120. Reduce request frequency."
          runbook: "docs/RUNBOOK.md#hitting-rate-limits"

      - alert: LinearHighRetryRate
        expr: |
          (
            sum(rate(linear_retries_total[10m]))
            /
            sum(rate(linear_requests_total[10m]))
          ) > 0.10
        for: 10m
        labels:
          severity: P3
          component: linear_client
        annotations:
          summary: "Linear retry rate > 10%"
          description: "{{ $value | humanizePercentage }} of requests are being retried. Check for API degradation."

      - alert: LinearHighLatency
        expr: |
          histogram_quantile(0.95,
            rate(linear_request_duration_seconds_bucket[5m])
          ) > 5
        for: 10m
        labels:
          severity: P3
          component: linear_client
        annotations:
          summary: "Linear API p95 latency > 5s"
          description: "p95 latency is {{ $value }}s. Check Linear status or retry storms."
```

---

## Grafana Dashboards

### Dashboard 1: Linear API Overview

**Panels:**

1. **Request Rate (Gauge)**
   ```promql
   sum(rate(linear_requests_total[5m]))
   ```

2. **Error Rate (Graph)**
   ```promql
   sum by(status_code) (rate(linear_requests_total{status_code=~"[45].."}[5m]))
   ```

3. **Latency (Graph - p50, p95, p99)**
   ```promql
   histogram_quantile(0.50, rate(linear_request_duration_seconds_bucket[5m]))
   histogram_quantile(0.95, rate(linear_request_duration_seconds_bucket[5m]))
   histogram_quantile(0.99, rate(linear_request_duration_seconds_bucket[5m]))
   ```

4. **Success vs Errors (Pie Chart)**
   ```promql
   sum(rate(linear_requests_total{status_code=~"2.."}[5m]))  # Success
   sum(rate(linear_requests_total{status_code=~"[45].."}[5m]))  # Errors
   ```

### Dashboard 2: Rate Limiting

**Panels:**

1. **Requests Remaining (Gauge)**
   ```promql
   linear_rate_limit_remaining{limit_type="requests"}
   ```

2. **Complexity Remaining (Gauge)**
   ```promql
   linear_rate_limit_remaining{limit_type="complexity"}
   ```

3. **Rate Limit Hit Count (Graph)**
   ```promql
   sum(rate(linear_retries_total{reason="rate_limited"}[5m]))
   ```

4. **Rate Limit Utilization % (Graph)**
   ```promql
   (1 - (linear_rate_limit_remaining{limit_type="requests"} / 120)) * 100
   ```

### Dashboard 3: Reliability

**Panels:**

1. **Retry Rate % (Graph)**
   ```promql
   (sum(rate(linear_retries_total[5m])) / sum(rate(linear_requests_total[5m]))) * 100
   ```

2. **Retries by Reason (Stacked Graph)**
   ```promql
   sum by(reason) (rate(linear_retries_total[5m]))
   ```

3. **Success After Retry (Stat)**
   ```promql
   # Requests that succeeded after retry
   sum(rate(linear_requests_total{status_code="200"}[5m]))
   - sum(rate(linear_requests_total[5m]) - sum(rate(linear_retries_total[5m])))
   ```

---

## Service Level Objectives (SLOs)

### Availability SLO: 99.9%

**Error Budget:** 43 minutes downtime per month

**Measurement:**
```promql
# Availability %
(
  sum(rate(linear_requests_total{status_code=~"2.."}[30d]))
  /
  sum(rate(linear_requests_total[30d]))
) * 100
```

**Burn Rate Alert:**
```yaml
- alert: LinearErrorBudgetBurning
  expr: |
    (
      1 - (
        sum(rate(linear_requests_total{status_code=~"2.."}[1h]))
        /
        sum(rate(linear_requests_total[1h]))
      )
    ) > 0.001  # 0.1% error rate (10x SLO)
  for: 5m
  labels:
    severity: P2
  annotations:
    summary: "Burning error budget 10x faster than SLO"
```

### Latency SLO: p95 < 2s

**Measurement:**
```promql
histogram_quantile(0.95, rate(linear_request_duration_seconds_bucket[5m]))
```

**Alert:**
```yaml
- alert: LinearLatencySLOViolation
  expr: |
    histogram_quantile(0.95,
      rate(linear_request_duration_seconds_bucket[5m])
    ) > 2
  for: 10m
  labels:
    severity: P3
  annotations:
    summary: "p95 latency violating SLO (> 2s)"
```

---

## Debugging with Metrics

### Find Slowest Operations

```promql
topk(5,
  histogram_quantile(0.95,
    sum by(operation) (rate(linear_request_duration_seconds_bucket[5m]))
  )
)
```

### Find Most Error-Prone Operations

```promql
topk(5,
  sum by(operation, error_type) (rate(linear_errors_total[5m]))
)
```

### Correlate Errors with Retries

```promql
# Operations with high error rate AND high retry rate
sum by(operation) (rate(linear_errors_total[5m]))
and on(operation)
sum by(operation) (rate(linear_retries_total[5m])) > 0.1
```

---

## Metric Retention

**Recommended:**
- Raw metrics: 15 days
- Aggregated (5m): 90 days
- Aggregated (1h): 1 year

**Storage estimate:**
- 5 metric families × 3 cardinality × 15s scrape = ~300 KB/day

---

## Integration Examples

### Prometheus Scrape Config

```yaml
scrape_configs:
  - job_name: 'go-linear'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:2112']
```

### Datadog Integration

```go
// Use Datadog's Prometheus scraper
// or convert metrics via datadog-agent
```

### CloudWatch Integration

```go
// Use CloudWatch's Prometheus remote write
// or AWS Distro for OpenTelemetry
```

---

## Resources

- **Prometheus Documentation:** https://prometheus.io/docs/
- **PromQL Basics:** https://prometheus.io/docs/prometheus/latest/querying/basics/
- **Grafana Dashboards:** https://grafana.com/docs/grafana/latest/dashboards/
- **Example Integration:** [examples/prometheus/main.go](../examples/prometheus/main.go)

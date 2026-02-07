# Prometheus Metrics

This API exposes Prometheus-compatible metrics for monitoring and observability.

## Metrics Endpoint

The metrics are exposed at `/metrics` endpoint:

```bash
curl http://localhost:8080/metrics
```

## Available Metrics

### HTTP Request Metrics

- **`http_request_duration_seconds`** (Histogram)
  - Duration of HTTP requests in seconds
  - Labels: `method`, `path`, `status`
  - Example: `http_request_duration_seconds{method="GET",path="/api/v1/secrets",status="200"}`

- **`http_requests_total`** (Counter)
  - Total number of HTTP requests
  - Labels: `method`, `path`, `status`
  - Example: `http_requests_total{method="POST",path="/api/v1/secrets",status="201"}`

- **`http_requests_in_flight`** (Gauge)
  - Number of HTTP requests currently being processed
  - Example: `http_requests_in_flight`

### MongoDB Operation Metrics

- **`mongodb_operation_duration_seconds`** (Histogram)
  - Duration of MongoDB operations in seconds
  - Labels: `operation`, `collection`
  - Example: `mongodb_operation_duration_seconds{operation="find",collection="secrets"}`

- **`mongodb_operations_total`** (Counter)
  - Total number of MongoDB operations
  - Labels: `operation`, `collection`, `status`
  - Example: `mongodb_operations_total{operation="insert",collection="secrets",status="success"}`

### Connection Metrics

- **`active_connections`** (Gauge)
  - Number of active connections
  - Example: `active_connections`

## Prometheus Configuration

To scrape metrics from this service, add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'simple-vault-api'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
        labels:
          service: 'simple-vault-api'
```

## Grafana Dashboard

You can create Grafana dashboards using these metrics. Common visualizations include:

- Request rate (requests per second)
- Request latency (p50, p95, p99)
- Error rate (4xx, 5xx responses)
- Active connections
- MongoDB operation performance

## Example Queries

### Request Rate
```promql
rate(http_requests_total[5m])
```

### Average Request Duration
```promql
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

### Error Rate
```promql
rate(http_requests_total{status=~"5.."}[5m])
```

### 95th Percentile Latency
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```


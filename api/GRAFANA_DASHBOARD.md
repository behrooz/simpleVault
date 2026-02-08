# Grafana Dashboard for Simple Vault API

This directory contains a pre-configured Grafana dashboard for monitoring the Simple Vault API metrics.

## Dashboard Features

The dashboard includes the following panels:

### HTTP Metrics
1. **HTTP Request Rate** - Shows requests per second by method, path, and status code
2. **In-Flight Requests** - Current number of requests being processed
3. **HTTP Request Latency** - p50, p95, and p99 percentiles
4. **HTTP Requests by Status Code** - Pie chart showing distribution of status codes
5. **HTTP Requests by Method** - Pie chart showing GET, POST, PUT, DELETE distribution
6. **HTTP Requests by Path** - Stacked area chart showing requests per endpoint
7. **Top 10 Endpoints by Request Rate** - Table showing most active endpoints
8. **Error Rate (5xx)** - Percentage of 5xx errors
9. **Overall p95 Latency** - Overall 95th percentile latency

### MongoDB Metrics
1. **MongoDB Operation Rate** - Operations per second by operation type and collection
2. **MongoDB Operation Latency** - p50, p95, and p99 percentiles for database operations
3. **MongoDB Operations by Status** - Success vs error distribution

### Connection Metrics
1. **Active Connections** - Current number of active connections

## Importing the Dashboard

### Method 1: Import via Grafana UI

1. Open Grafana in your browser
2. Click on the **"+"** icon in the left sidebar
3. Select **"Import dashboard"**
4. Click **"Upload JSON file"** and select `grafana-dashboard.json`
   - OR paste the JSON content directly into the text area
5. Select your Prometheus data source from the dropdown
6. Click **"Import"**

### Method 2: Import via API

```bash
# Set your Grafana credentials
GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASSWORD="admin"

# Import the dashboard
curl -X POST \
  -H "Content-Type: application/json" \
  -d @grafana-dashboard.json \
  "${GRAFANA_URL}/api/dashboards/db" \
  -u "${GRAFANA_USER}:${GRAFANA_PASSWORD}"
```

### Method 3: Import via Grafana CLI

```bash
grafana-cli admin import-dashboard grafana-dashboard.json
```

## Prerequisites

1. **Prometheus Data Source**: Make sure you have a Prometheus data source configured in Grafana
   - The dashboard uses `${DS_PROMETHEUS}` variable, which will prompt you to select a data source on import
   - If you want to hardcode a data source, replace `${DS_PROMETHEUS}` with your data source UID

2. **Prometheus Scraping**: Ensure Prometheus is configured to scrape metrics from your API:
   ```yaml
   scrape_configs:
     - job_name: 'simple-vault-api'
       scrape_interval: 15s
       static_configs:
         - targets: ['localhost:8080']  # Update with your API endpoint
   ```

3. **Metrics Endpoint**: Verify the `/metrics` endpoint is accessible:
   ```bash
   curl http://localhost:8080/metrics
   ```

## Customization

### Changing the Data Source

If you want to use a specific Prometheus data source:

1. Open the dashboard JSON file
2. Find all occurrences of `"uid": "${DS_PROMETHEUS}"`
3. Replace with your Prometheus data source UID (found in Grafana → Configuration → Data Sources)

### Adjusting Time Ranges

The default time range is set to "Last 6 hours". To change:

1. Edit the `"time"` section in the JSON:
   ```json
   "time": {
     "from": "now-24h",  // Change to desired range
     "to": "now"
   }
   ```

### Modifying Refresh Interval

The dashboard refreshes every 30 seconds by default. To change:

1. Edit the `"refresh"` field:
   ```json
   "refresh": "1m"  // Options: 5s, 10s, 30s, 1m, 5m, 15m, 30m, 1h
   ```

## Troubleshooting

### No Data Showing

1. **Check Prometheus**: Verify Prometheus is scraping your API
   ```bash
   # Check Prometheus targets
   curl http://localhost:9090/api/v1/targets
   ```

2. **Check Metrics Endpoint**: Verify metrics are being exposed
   ```bash
   curl http://localhost:8080/metrics | grep http_requests_total
   ```

3. **Check Time Range**: Make sure your time range includes when data was collected

4. **Check Data Source**: Verify the Prometheus data source is correctly configured in Grafana

### Panels Showing "No Data"

- Some panels (like MongoDB metrics) require the metrics to be actively recorded
- If you haven't instrumented MongoDB operations yet, those panels will show no data
- HTTP metrics should work immediately as they're automatically collected

## Dashboard Variables (Optional)

You can add template variables for filtering. Example:

```json
"templating": {
  "list": [
    {
      "name": "method",
      "type": "query",
      "datasource": {
        "type": "prometheus",
        "uid": "${DS_PROMETHEUS}"
      },
      "query": "label_values(http_requests_total, method)",
      "current": {
        "selected": false,
        "text": "All",
        "value": "$__all"
      },
      "includeAll": true
    }
  ]
}
```

Then use `$method` in your queries to filter by HTTP method.

## Support

For issues or questions:
- Check the [METRICS.md](./METRICS.md) file for metric documentation
- Verify your Prometheus configuration
- Check Grafana logs for errors


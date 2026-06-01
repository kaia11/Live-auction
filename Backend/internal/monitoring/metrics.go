package monitoring

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type Metrics struct {
	httpRequestsTotal      int64
	httpRequestDurationMS  int64
	bidRequestsTotal       int64
	bidSuccessTotal        int64
	bidFailureTotal        int64
	settlementsTotal       int64
	settlementDelayMSTotal int64
	wsOpenedTotal          int64
	wsClosedTotal          int64
	wsActiveConnections    int64
	mu                     sync.Mutex
	errorCounters          map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		errorCounters: make(map[string]int64),
	}
}

func (m *Metrics) RecordHTTPRequest(durationMS int64) {
	atomic.AddInt64(&m.httpRequestsTotal, 1)
	atomic.AddInt64(&m.httpRequestDurationMS, durationMS)
}

func (m *Metrics) RecordBidAttempt() {
	atomic.AddInt64(&m.bidRequestsTotal, 1)
}

func (m *Metrics) RecordBidSuccess() {
	atomic.AddInt64(&m.bidSuccessTotal, 1)
}

func (m *Metrics) RecordBidFailure(reason string) {
	atomic.AddInt64(&m.bidFailureTotal, 1)
	m.RecordError("bid_" + sanitizeMetricLabel(reason))
}

func (m *Metrics) RecordSettlement(delayMS int64) {
	atomic.AddInt64(&m.settlementsTotal, 1)
	if delayMS > 0 {
		atomic.AddInt64(&m.settlementDelayMSTotal, delayMS)
	}
}

func (m *Metrics) RecordError(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCounters[sanitizeMetricLabel(name)]++
}

func (m *Metrics) RecordWSConnect() {
	atomic.AddInt64(&m.wsOpenedTotal, 1)
	atomic.AddInt64(&m.wsActiveConnections, 1)
}

func (m *Metrics) RecordWSDisconnect() {
	atomic.AddInt64(&m.wsClosedTotal, 1)
	atomic.AddInt64(&m.wsActiveConnections, -1)
}

func (m *Metrics) RenderPrometheus() string {
	var b strings.Builder

	writeMetric(&b, "auction_http_requests_total", atomic.LoadInt64(&m.httpRequestsTotal), "Total HTTP requests.")
	writeMetric(&b, "auction_http_request_duration_ms_total", atomic.LoadInt64(&m.httpRequestDurationMS), "Accumulated HTTP request duration in milliseconds.")
	writeMetric(&b, "auction_bid_requests_total", atomic.LoadInt64(&m.bidRequestsTotal), "Total bid create attempts.")
	writeMetric(&b, "auction_bid_success_total", atomic.LoadInt64(&m.bidSuccessTotal), "Total successful bids.")
	writeMetric(&b, "auction_bid_failure_total", atomic.LoadInt64(&m.bidFailureTotal), "Total failed bids.")
	writeMetric(&b, "auction_settlements_total", atomic.LoadInt64(&m.settlementsTotal), "Total settled sessions.")
	writeMetric(&b, "auction_settlement_delay_ms_total", atomic.LoadInt64(&m.settlementDelayMSTotal), "Accumulated settlement delay in milliseconds.")
	writeMetric(&b, "auction_ws_connections_opened_total", atomic.LoadInt64(&m.wsOpenedTotal), "Total opened websocket connections.")
	writeMetric(&b, "auction_ws_connections_closed_total", atomic.LoadInt64(&m.wsClosedTotal), "Total closed websocket connections.")
	writeGaugeMetric(&b, "auction_ws_active_connections", atomic.LoadInt64(&m.wsActiveConnections), "Current active websocket connections.")

	m.mu.Lock()
	keys := make([]string, 0, len(m.errorCounters))
	for key := range m.errorCounters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(&b, "auction_errors_total{type=%q} %d\n", key, m.errorCounters[key])
	}
	m.mu.Unlock()

	return b.String()
}

func writeMetric(b *strings.Builder, name string, value int64, help string) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s counter\n", name)
	fmt.Fprintf(b, "%s %d\n", name, value)
}

func writeGaugeMetric(b *strings.Builder, name string, value int64, help string) {
	fmt.Fprintf(b, "# HELP %s %s\n", name, help)
	fmt.Fprintf(b, "# TYPE %s gauge\n", name)
	fmt.Fprintf(b, "%s %d\n", name, value)
}

func sanitizeMetricLabel(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "unknown"
	}
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

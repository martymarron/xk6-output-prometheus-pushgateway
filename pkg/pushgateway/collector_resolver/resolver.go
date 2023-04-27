package pushgateway

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.k6.io/k6/metrics"
)

// CollectorResolver is an interface to resolve the various types of the [metrics.Metric]
// to the [prometheus.Collector].
// Respective [k6 metric type] are solved by following the [conversion rule] which
// the [xk6-output-prometheus-remote] extension applies.
//
// [k6 metric type]: https://k6.io/docs/using-k6/metrics/#metric-types
// [conversion rule]: https://k6.io/blog/k6-loves-prometheus/#mapping-k6-metrics-types
// [xk6-output-prometheus-remote]: https://github.com/grafana/xk6-output-prometheus-remote
type CollectorResolver func(sample metrics.Sample, labels prometheus.Labels, prefix string) []prometheus.Collector

// CreateResolveer is a factory method to create the [ColloectorResolver] implementation
// corresponding to the given [k6 metric type].
//
// Example use case:
//
//	collectorResolver := collector_resolver.CreateResolver(sample.Metric.Type)
//	collectors := collectorResolver(sample.Metric, time.Now())
//
// [k6 metric type]: https://k6.io/docs/using-k6/metrics/#metric-types
func CreateResolver(t metrics.MetricType) CollectorResolver {
	var resolver CollectorResolver
	switch t {
	case metrics.Counter:
		resolver = resolveCounter
	case metrics.Gauge:
		resolver = resolveGauge
	case metrics.Rate:
		resolver = resolveRate
	case metrics.Trend:
		resolver = resolveTrend
	}
	return resolver
}

func resolveCounter(sample metrics.Sample, labels prometheus.Labels, prefix string) []prometheus.Collector {
	counter := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name:        getPrefixedName(prefix, sample.Metric.Name),
			ConstLabels: labels,
		},
		func() float64 {
			counterSink := sample.Metric.Sink.(*metrics.CounterSink)
			return sample.Metric.Sink.Format(sample.GetTime().Sub(counterSink.First))["rate"]
		},
	)
	return []prometheus.Collector{counter}
}

func resolveGauge(sample metrics.Sample, labels prometheus.Labels, prefix string) []prometheus.Collector {
	gauge := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        getPrefixedName(prefix, sample.Metric.Name),
			ConstLabels: labels,
		},
		func() float64 {
			return sample.Value
		},
	)
	return []prometheus.Collector{gauge}
}

func resolveRate(sample metrics.Sample, labels prometheus.Labels, prefix string) []prometheus.Collector {
	gauge := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        getPrefixedName(prefix, sample.Metric.Name),
			ConstLabels: labels,
		},
		func() float64 {
			return sample.Value
		},
	)
	return []prometheus.Collector{gauge}
}

func resolveTrend(sample metrics.Sample, labels prometheus.Labels, prefix string) []prometheus.Collector {
	sink := sample.Metric.Sink.Format(time.Duration(0))

	collectors := make([]prometheus.Collector, 0)
	for k, v := range sink {
		// Remove "(" and ")" from the name of prometheus collector
		// Becuase these are not acceptable as collector name.
		suffix := strings.ReplaceAll(strings.ReplaceAll(k, "(", ""), ")", "")

		name := fmt.Sprintf("%s_%s", sample.Metric.Name, suffix)
		gauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:        getPrefixedName(prefix, name),
				ConstLabels: labels,
			},
		)
		gauge.Set(v)
		collectors = append(collectors, gauge)
	}
	return collectors
}

func getPrefixedName(prefix string, name string) string {
	if prefix == "" {
		return name
	}

	return fmt.Sprintf("%s_%s", prefix, name)
}

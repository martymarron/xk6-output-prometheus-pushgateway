package pushgateway

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.k6.io/k6/metrics"
)

func TestCreateResolver(t *testing.T) {
	fixture := []struct {
		input    metrics.MetricType
		resolver CollectorResolver
	}{
		{metrics.Counter, resolveCounter},
		{metrics.Gauge, resolveGauge},
		{metrics.Rate, resolveRate},
		{metrics.Trend, resolveTrend},
	}

	for _, v := range fixture {
		resolver := CreateResolver(v.input)
		p1 := fmt.Sprintf("%v", resolver)
		p2 := fmt.Sprintf("%v", v.resolver)
		name := fmt.Sprintf("Expected %v, but %v is returned.", p1, p2)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if p1 != p2 {
				t.Errorf(name)
			}
		})
	}
}

func TestResolveCounter(t *testing.T) {
	// Given
	timeFirst := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	timeSecond := time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC)
	metricName := "sample_metric"
	metricValue := 100
	metricSeed, _ := metrics.NewRegistry().NewMetric(metricName, metrics.Counter)
	sink := &metrics.CounterSink{
		Value: float64(metricValue),
		First: timeFirst,
	}
	metric := &metrics.Metric{
		Name:       metricSeed.Name,
		Type:       metricSeed.Type,
		Contains:   metricSeed.Contains,
		Tainted:    metricSeed.Tainted,
		Thresholds: metricSeed.Thresholds,
		Submetrics: metricSeed.Submetrics,
		Sub:        metricSeed.Sub,
		Sink:       sink,
		Observed:   metricSeed.Observed,
	}
	sample := metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: metric,
		},
		Time:  timeSecond,
		Value: sink.Value,
	}

	// When
	collectors := resolveCounter(sample, nil, "")

	// Then
	if len(collectors) != 1 {
		t.Errorf("Length of returned collector slice should be one. Actual: %d", len(collectors))
	}
	reg := prometheus.NewRegistry()
	reg.Register(collectors[0])
	mfs, _ := prometheus.Gatherers{reg}.Gather()
	if len(mfs) != 1 {
		t.Errorf("Length of MetricFamily should be one. Actual: %+v", mfs)
		return
	}
	mType := metrics.MetricType(mfs[0].GetType())
	mName := mfs[0].GetName()
	if mType != metrics.Counter {
		t.Errorf("Metrics type should be counter. Actual: %s", mType)
	}
	if mName != metricName {
		t.Errorf("Metrics name should be %s. Actual: %s", metricName, mName)
	}
	ms := mfs[0].GetMetric()
	if len(ms) != 1 {
		t.Errorf("Length of MetricFamily should be one. Actual: %+v", ms)
	}

	m := ms[0]
	if int(m.GetCounter().GetValue()) != metricValue {
		t.Errorf("Metric value should be %d. Actual: %+v", metricValue, m)
	}
}

func TestResolveGauge(t *testing.T) {
	// Given
	metricName := "sample_metric"
	metricValue := 100.00
	metricSeed, _ := metrics.NewRegistry().NewMetric(metricName, metrics.Gauge)
	sink := &metrics.GaugeSink{}
	sink.Add(metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: metricSeed,
		},
		Time:  time.Now(),
		Value: metricValue,
	})
	metric := &metrics.Metric{
		Name:       metricSeed.Name,
		Type:       metricSeed.Type,
		Contains:   metricSeed.Contains,
		Tainted:    metricSeed.Tainted,
		Thresholds: metricSeed.Thresholds,
		Submetrics: metricSeed.Submetrics,
		Sub:        metricSeed.Sub,
		Sink:       sink,
		Observed:   metricSeed.Observed,
	}
	sample := metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: metric,
		},
		Time:  time.Now(),
		Value: sink.Value,
	}

	// When
	collectors := resolveGauge(sample, nil, "")

	// Then
	if len(collectors) != 1 {
		t.Errorf("Length of returned collector slice should be one. Actual: %d", len(collectors))
	}
	reg := prometheus.NewRegistry()
	reg.Register(collectors[0])
	mfs, _ := prometheus.Gatherers{reg}.Gather()
	if len(mfs) != 1 {
		t.Errorf("Length of MetricFamily should be one. Actual: %+v", mfs)
		return
	}
	mType := metrics.MetricType(mfs[0].GetType())
	mName := mfs[0].GetName()
	if mType != metrics.Gauge {
		t.Errorf("Metrics type should be gauge. Actual: %s", mType)
	}
	if mName != metricName {
		t.Errorf("Metrics name should be %s. Actual: %s", metricName, mName)
	}
	ms := mfs[0].GetMetric()
	if len(ms) != 1 {
		t.Errorf("Length of MetricFamily should be one. Actual: %+v", ms)
	}

	m := ms[0]
	if m.GetGauge().GetValue() != metricValue {
		t.Errorf("Metric value should be %f. Actual: %+v", metricValue, m)
	}
}

func TestResolveRate(t *testing.T) {
	// Given
	metricName := "sample_metric"
	metricSeed, _ := metrics.NewRegistry().NewMetric(metricName, metrics.Rate)
	trues := int64(50)
	total := int64(100)
	sink := &metrics.RateSink{
		Trues: trues,
		Total: total,
	}
	metric := &metrics.Metric{
		Name:       metricSeed.Name,
		Type:       metricSeed.Type,
		Contains:   metricSeed.Contains,
		Tainted:    metricSeed.Tainted,
		Thresholds: metricSeed.Thresholds,
		Submetrics: metricSeed.Submetrics,
		Sub:        metricSeed.Sub,
		Sink:       sink,
		Observed:   metricSeed.Observed,
	}
	sample := metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: metric,
		},
		Time:  time.Now(),
		Value: float64(sink.Trues) / float64(sink.Total),
	}

	// When
	collectors := resolveRate(sample, nil, "")

	// Then
	if len(collectors) != 1 {
		t.Errorf("Length of returned collector slice should be one. Actual: %d", len(collectors))
	}
	reg := prometheus.NewRegistry()
	reg.Register(collectors[0])
	mfs, _ := prometheus.Gatherers{reg}.Gather()
	if len(mfs) != 1 {
		t.Errorf("Length of MetricFamily should be one. Actual: %+v", mfs)
		return
	}
	mType := metrics.MetricType(mfs[0].GetType())
	mName := mfs[0].GetName()
	if mType != metrics.Gauge {
		t.Errorf("Metrics type should be gauge. Actual: %s", mType)
	}
	if mName != metricName {
		t.Errorf("Metrics name should be %s. Actual: %s", metricName, mName)
	}
	ms := mfs[0].GetMetric()
	if len(ms) != 1 {
		t.Errorf("Length of MetricFamily should be one. Actual: %+v", ms)
	}

	m := ms[0]
	metricValue := float64(trues) / float64(total)
	if m.GetGauge().GetValue() != metricValue {
		t.Errorf("Metric value should be %f. Actual: %+v", metricValue, m)
	}
}

func TestResolveTrent(t *testing.T) {
	// Given
	metricName := "sample_metric"
	metricSeed, _ := metrics.NewRegistry().NewMetric(metricName, metrics.Trend)
	min, max := 10.00, 90.00
	values := []float64{50.00, min, max}
	sum, avg := 150.00, 50.00
	sink := &metrics.TrendSink{
		Values: values,
		Count:  uint64(len(values)),
		Min:    min,
		Max:    max,
		Sum:    sum,
		Avg:    avg,
	}
	metric := &metrics.Metric{
		Name:       metricSeed.Name,
		Type:       metricSeed.Type,
		Contains:   metricSeed.Contains,
		Tainted:    metricSeed.Tainted,
		Thresholds: metricSeed.Thresholds,
		Submetrics: metricSeed.Submetrics,
		Sub:        metricSeed.Sub,
		Sink:       sink,
		Observed:   metricSeed.Observed,
	}
	sample := metrics.Sample{
		TimeSeries: metrics.TimeSeries{
			Metric: metric,
		},
		Time:  time.Now(),
		Value: float64(sink.Count),
	}

	// When
	collectors := resolveTrend(sample, nil, "")

	// Then
	if len(collectors) != 6 {
		t.Errorf("Length of returned collector slice should be one. Actual: %d", len(collectors))
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors...)
	mfs, _ := prometheus.Gatherers{reg}.Gather()
	if len(mfs) != 6 {
		t.Errorf("Length of MetricFamily should be one. Actual: %+v", mfs)
		return
	}

	for _, mf := range mfs {
		mType := metrics.MetricType(mf.GetType())
		mName := mf.GetName()
		t.Run(mName, func(t *testing.T) {
			if mType != metrics.Gauge {
				t.Errorf("Metrics type should be gauge. Actual: %s", mType)
			}
			if !strings.HasPrefix(mName, metricName) {
				t.Errorf("Metrics name should start with %s. Actual: %s", metricName, mName)
			}

			for _, m := range mf.GetMetric() {
				t.Logf("Metric %s's value is %f.", mName, *m.GetGauge().Value)
			}
		})
	}
}

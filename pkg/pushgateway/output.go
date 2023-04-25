package pushgateway

import (
	"fmt"
	"sync"
	"time"

	collector_resolver "github.com/martymarron/xk6-output-prometheus-pushgateway/pkg/pushgateway/collector_resolver"

	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
)

// Output implements the lib.Output interface
type Output struct {
	output.SampleBuffer

	config          Config
	periodicFlusher *output.PeriodicFlusher
	logger          logrus.FieldLogger
	waitGroup       sync.WaitGroup
}

var _ output.Output = new(Output)

// New creates an instance of the collector
func New(p output.Params) (*Output, error) {
	conf, err := NewConfig(p)
	if err != nil {
		return nil, err
	}
	// Some setupping code

	return &Output{
		config: conf,
		logger: p.Logger,
	}, nil
}

func (o *Output) Description() string {
	return fmt.Sprintf("pushgateway: %s, job: %s, labels: %s", o.config.PushGWUrl, o.config.JobName, o.config.Labels)
}

func (o *Output) Stop() error {
	o.logger.Debug("Stopping...")
	defer o.logger.Debug("Stopped!")
	o.periodicFlusher.Stop()
	o.waitGroup.Wait()

	return nil
}

func (o *Output) Start() error {
	o.logger.Debug("Starting...")

	// Here we should connect to a service, open a file or w/e else we decided we need to do

	pf, err := output.NewPeriodicFlusher(o.config.PushInterval, o.flushMetrics)
	if err != nil {
		return err
	}
	o.logger.Debug("Started!")
	o.periodicFlusher = pf

	return nil
}

// GetLabels gets labels from the Config
func (o *Output) GetLabels() prometheus.Labels {
	return o.config.Labels
}

func (o *Output) flushMetrics() {
	sampleContainers := o.GetBufferedSamples()

	o.waitGroup.Add(1)
	go func() {
		defer o.waitGroup.Done()

		sampleMap := extractPushSamples(sampleContainers)
		o.logger.WithFields(dumpk6Sample(sampleMap)).Debug("Dump k6 samples.")
		labels := o.GetLabels()
		collectors := o.convertk6SamplesToPromCollectors(sampleMap, labels)

		registry := prometheus.NewPedanticRegistry()
		registry.MustRegister(collectors...)
		o.logger.WithFields(dumpPrometheusCollector(registry)).Debug("Dump collectors.")

		pusher := push.New(o.config.PushGWUrl, o.config.JobName)
		if err := pusher.Gatherer(registry).Add(); err != nil {
			o.logger.
				Errorf("Could not add to Pushgateway: %v", err)
		}
	}()
}

func extractPushSamples(sampleContainers []metrics.SampleContainer) map[string]metrics.Sample {
	// To avoid duplicated metric registration,
	// store metric name and its value as a map,
	// and overwrite the value by the latest one.
	sampleMap := make(map[string]metrics.Sample)
	for _, sampleContainer := range sampleContainers {
		samples := sampleContainer.GetSamples()
		for _, sample := range samples {
			key := sample.Metric.Name
			sampleMap[key] = sample
		}
	}
	return sampleMap
}

func (o *Output) convertk6SamplesToPromCollectors(samplesMap map[string]metrics.Sample, labels prometheus.Labels) []prometheus.Collector {
	collectors := make([]prometheus.Collector, 0)
	for _, sample := range samplesMap {
		resolver := collector_resolver.CreateResolver(sample.Metric.Type)
		collectors = append(collectors, resolver(sample, labels, o.config.Namespace)...)
	}
	return collectors
}

func dumpk6Sample(samplesMap map[string]metrics.Sample) logrus.Fields {
	var value float64
	t := time.Since(time.Now())
	fields := logrus.Fields{}
	for _, sample := range samplesMap {
		switch sample.Metric.Type {
		case metrics.Counter:
			value = sample.Metric.Sink.Format(t)["count"]
		case metrics.Gauge:
			value = sample.Metric.Sink.Format(t)["value"]
		case metrics.Rate:
			value = sample.Metric.Sink.Format(t)["rate"]
		}
		fields[sample.Metric.Name] = map[string]interface{}{
			"sample_value": sample.Value,
			"sink_value":   value,
			"name":         sample.Metric.Name,
			"type":         sample.Metric.Type,
		}
	}
	return fields
}

func dumpPrometheusCollector(reg *prometheus.Registry) logrus.Fields {
	metricFamilies, _ := prometheus.Gatherers{reg}.Gather()
	fields := logrus.Fields{}
	for _, metricFamily := range metricFamilies {
		fields[metricFamily.GetName()] = metricFamily.String()
	}
	return fields
}

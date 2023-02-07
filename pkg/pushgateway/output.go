package pushgateway

import (
	"time"

	collector_resolver "github.com/martymarron/xk6-output-prometheus-pushgateway/pkg/pushgateway/collector_resolver"

	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.k6.io/k6/output"
)

// Output implements the lib.Output interface
type Output struct {
	output.SampleBuffer

	config          Config
	periodicFlusher *output.PeriodicFlusher
	logger          logrus.FieldLogger
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
	return "pushgateway: " + o.config.PushGWUrl
}

func (o *Output) Stop() error {
	o.logger.Debug("Stopping...")
	defer o.logger.Debug("Stopped!")
	o.periodicFlusher.Stop()
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

func (o *Output) flushMetrics() {
	samples := o.GetBufferedSamples()
	start := time.Now()
	var count int

	for _, sc := range samples {
		pusher := push.New(o.config.PushGWUrl, o.config.JobName)
		samples := sc.GetSamples()
		count += len(samples)
		for _, sample := range samples {
			collectorResolver := collector_resolver.CreateResolver(sample.Metric.Type)
			collectors := collectorResolver(sample)
			pushCollectors(collectors, pusher)
		}
		if err := pusher.Add(); err != nil {
			o.logger.WithError(err).Debug("Could not add to Pushgateway")
		}
	}

	if count > 0 {
		o.logger.WithField("t", time.Since(start)).WithField("count", count).Debug("Wrote metrics to stdout")
	}
}

func pushCollectors(cs []prometheus.Collector, pusher *push.Pusher) {
	for _, collector := range cs {
		pusher.Collector(collector)
	}
}

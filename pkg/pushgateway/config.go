package pushgateway

import (
	"fmt"
	"time"

	"go.k6.io/k6/output"
)

const (
	defaultPushGWUrl    = "http://localhost:9091"
	defaultPushInterval = 10 * time.Second
	defaultJobName      = "k6_load_testing"
)

// Config is the config for the template collector
type Config struct {
	PushGWUrl    string
	PushInterval time.Duration
	JobName      string
}

// NewConfig creates a new Config instance from the provided output.Params
func NewConfig(p output.Params) (Config, error) {
	cfg := Config{
		PushGWUrl:    defaultPushGWUrl,
		PushInterval: defaultPushInterval,
		JobName:      defaultJobName,
	}

	for k, v := range p.Environment {
		switch k {
		case "K6_PUSH_INTERVAL":
			var err error
			cfg.PushInterval, err = time.ParseDuration(v)
			if err != nil {
				return cfg, fmt.Errorf("error parsing environment variable 'K6_TEMPLATE_PUSH_INTERVAL': %w", err)
			}
		case "K6_PUSHGATEWAY_URL":
			cfg.PushGWUrl = v
		case "K6_JOB_NAME":
			cfg.JobName = v
		}
	}
	return cfg, nil
}

package pushgateway

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"go.k6.io/k6/output"
)

const (
	defaultPushGWUrl    = "http://localhost:9091"
	defaultPushInterval = 10 * time.Second
	defaultJobName      = "k6_load_testing"
)

// Config is the config for the template collector
type Config struct {
	JobName      string
	Labels       map[string]string
	PushGWUrl    string
	PushInterval time.Duration
}

// NewConfig creates a new Config instance from the provided output.Params
func NewConfig(p output.Params) (Config, error) {
	cfg := Config{
		JobName:      defaultJobName,
		Labels:       map[string]string{},
		PushGWUrl:    defaultPushGWUrl,
		PushInterval: defaultPushInterval,
	}

	if val, ok := p.ScriptOptions.External["pushgateway"]; ok {
		err := json.Unmarshal(val, &cfg.Labels)
		if err != nil {
			j, err := json.Marshal(&val)
			if err != nil {
				return cfg, errors.Wrap(err, fmt.Sprintf(
					"unable to get labels for JSON options.ext.pushgateway dictionary %s", string(j)))
			} else {
				return cfg, errors.Wrap(err, "unable to get labels for JSON options.ext.pushgateway dictionary")
			}

		}
		p.Logger.Debugf("Pushgateway labels from JSON options.ext.pushgateway dictionary %+v", cfg.Labels)
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
		if strings.HasPrefix(k, "K6_LABEL_") {
			key := strings.ToLower(k[9:])
			cfg.Labels[key] = strings.ToLower(v)
		}
	}
	p.Logger.Debugf("Pushgateway labels %+v", cfg.Labels)
	return cfg, nil
}

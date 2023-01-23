package pushgateway

import (
	"encoding/json"
	"fmt"
	"strings"
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
	Debug        bool
	JobName      string
	Labels       map[string]string
	PushGWUrl    string
	PushInterval time.Duration
}

// NewConfig creates a new Config instance from the provided output.Params
func NewConfig(p output.Params) (Config, error) {
	cfg := Config{
		Debug:        false,
		JobName:      defaultJobName,
		Labels:       map[string]string{},
		PushGWUrl:    defaultPushGWUrl,
		PushInterval: defaultPushInterval,
	}

	if _, ok := p.Environment["K6_DEBUG"]; ok {
		cfg.Debug = true
	}

	if val, ok := p.ScriptOptions.External["pushgateway"]; ok {
		err := json.Unmarshal(val, &cfg.Labels)
		if err != nil {
			j, err := json.Marshal(&val)
			if cfg.Debug {
				if err != nil {
					fmt.Printf("xk6-output-prometheus-pushgateway: WARN: "+
						"unable to get labels for JSON options.ext.pushgateway dictionary %s\n", string(j))
				} else {
					fmt.Printf("xk6-output-prometheus-pushgateway: WARN: " +
						"unable to get labels for JSON options.ext.pushgateway dictionary\n")
				}
			}
		}
		if cfg.Debug {
			fmt.Printf("Pushgateway labels from JSON options.ext.pushgateway dictionary %+v\n", cfg.Labels)
		}
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
	if cfg.Debug {
		fmt.Printf("Pushgateway labels %+v\n", cfg.Labels)
	}
	return cfg, nil
}

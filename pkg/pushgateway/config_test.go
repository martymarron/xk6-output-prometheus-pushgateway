package pushgateway_test

import (
	"testing"

	"github.com/martymarron/xk6-output-prometheus-pushgateway/pkg/pushgateway"
	"go.k6.io/k6/output"
)

func TestConfigLabels(t *testing.T) {
	p := output.Params{
		Environment: map[string]string{},
	}
	cfg, err := pushgateway.NewConfig(p)
	if err != nil {
		t.Errorf("Unable to create a new config, error: %v", err)
	}
	if len(cfg.Labels) != 0 {
		t.Errorf("Unexpecten labels value %+v", cfg)
	}

	p.Environment["K6_LABEL_ENV"] = "PROD"
	p.Environment["K6_LABEL_APP"] = "APP"
	cfg, err = pushgateway.NewConfig(p)
	if err != nil {
		t.Errorf("Unable to create a new config, error: %v", err)
	}
	if len(cfg.Labels) == 2 &&
		cfg.Labels["app"] == "APP" &&
		cfg.Labels["env"] == "PROD" {
		t.Errorf("Unexpecten labels value %+v", cfg)
	}
}

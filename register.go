package pushgateway

import (
	"github.com/martymarron/xk6-output-prometheus-pushgateway/pkg/pushgateway"

	"go.k6.io/k6/output"
)

func init() {
	name := "output-prometheus-pushgateway"
	output.RegisterExtension(name, func(p output.Params) (output.Output, error) {
		p.Logger = p.Logger.WithField("component", name)
		return pushgateway.New(p)
	})
}

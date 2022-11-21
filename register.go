package pushgateway

import (
	"github.com/martymarron/xk6-output-prometheus-pushgateway/pkg/pushgateway"

	"go.k6.io/k6/output"
)

func init() {
	output.RegisterExtension("output-prometheus-pushgateway", func(p output.Params) (output.Output, error) {
		return pushgateway.New(p)
	})
}

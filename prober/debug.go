package prober

import (
	"bytes"
	"fmt"

	"github.com/prometheus/blackbox_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func debugModule(buf *bytes.Buffer, name string, module *config.Module) {
	fmt.Fprintf(buf, "Module configuration:\n")

	data, err := module.MarshalIcinga()
	if err != nil {
		fmt.Fprintf(buf, "Error marshalling config: %s\n", err)
	}

	buf.Write(c)
	buf.Write('\n')
}

func debugRegistry(buf *bytes.Buffer, registry *prometheus.Registry) {
	fmt.Fprintf(buf, "Metrics that would have been returned:\n")

	mfs, err := registry.Gather()
	if err != nil {
		fmt.Fprintf(buf, "Error gathering metrics: %s\n", err)
	}

	for _, mf := range mfs {
		expfmt.MetricFamilyToText(buf, mf)
	}

	buf.Write('\n')
}

func debugLogger(buf *bytes.Buffer, logger *scrapeLogger) {
	fmt.Fprintf(buf, "Logs for the probe:\n")
	logger.buffer.WriteTo(buf)
	buf.Write('\n')
}

package prober

import (
	"bytes"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/config"
	monitoring "github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/nagios"
)

func debugModule(buf *bytes.Buffer, name string, module *config.Module) {
	fmt.Fprintf(buf, "Module configuration:\n")

	data, err := module.MarshalIcinga(name)
	if err != nil {
		fmt.Fprintf(buf, "Error marshalling config: %s\n", err)
	}

	buf.Write(data)
	buf.WriteByte('\n')
}

func debugPlugin(buf *bytes.Buffer, plugin *monitoring.Plugin, output *monitoring.PluginResult, err error) {
	fmt.Fprintf(buf, "Plugin execution:\n")

	fmt.Fprintf(buf, "Execv: %s\n", plugin)
	if err != nil {
		fmt.Fprintf(buf, "Error: %s\n", err)
	} else {
		if output.Error != nil {
			fmt.Fprintf(buf, "Error: %s\n", output.Error)
		}

		fmt.Fprintf(buf, "Output: %s\n", output)
	}
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

	buf.WriteByte('\n')
}

func debugLogger(buf *bytes.Buffer, logger *scrapeLogger) {
	fmt.Fprintf(buf, "Logs for the probe:\n")
	logger.buffer.WriteTo(buf)
	buf.WriteByte('\n')
}

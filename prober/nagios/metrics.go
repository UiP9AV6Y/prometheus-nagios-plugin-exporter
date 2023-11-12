package nagios

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/config"
	monitoring "github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/nagios"
)

type PluginMetrics struct {
	probeExitGauge     prometheus.Gauge
	probeSuccessGauge  prometheus.Gauge
	probeDurationGauge prometheus.Gauge
}

func NewPluginMetrics(module *config.Module, namespace string) *PluginMetrics {
	probeExitGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "probe_exit_code",
		Help:      "Probe command exit code",
	})
	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "probe_success",
		Help:      "Displays whether or not the probe was a success",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "probe_duration_seconds",
		Help:      "Returns how long the probe took to complete in seconds",
	})
	result := &PluginMetrics{
		probeExitGauge:     probeExitGauge,
		probeSuccessGauge:  probeSuccessGauge,
		probeDurationGauge: probeDurationGauge,
	}

	return result
}

func (m *PluginMetrics) Register(registry *prometheus.Registry) error {
	if err := registry.Register(m.probeExitGauge); err != nil {
		return err
	}

	if err := registry.Register(m.probeSuccessGauge); err != nil {
		return err
	}

	if err := registry.Register(m.probeDurationGauge); err != nil {
		return err
	}

	return nil
}

func (m *PluginMetrics) Report(output *monitoring.PluginResult, err error, duration float64) {
	m.probeDurationGauge.Set(duration)

	if err == nil {
		m.probeExitGauge.Set(float64(output.Status))
		m.probeSuccessGauge.Set(1)
	}
}

package prober

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/blackbox_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(w http.ResponseWriter, r *http.Request, c *config.Config, logger log.Logger, params url.Values,
	moduleUnknownCounter prometheus.Counter,
	logLevelProber level.Option) {

	if params == nil {
		params = r.URL.Query()
	}

	moduleName := params.Get("module")
	module, ok := c.Modules[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("Unknown module %q", moduleName), http.StatusBadRequest)
		level.Debug(logger).Log("msg", "Unknown module", "module", moduleName)
		if moduleUnknownCounter != nil {
			moduleUnknownCounter.Add(1)
		}
		return
	}

	target := params.Get("target")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	timeoutSeconds, err := getTimeout(r, module.Timeout)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse timeout from Prometheus header: %s", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds*float64(time.Second)))
	defer cancel()
	r = r.WithContext(ctx)

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_duration_seconds",
		Help: "Returns how long the probe took to complete in seconds",
	})

	sl := newScrapeLogger(logger, moduleName, target, logLevelProber)
	level.Info(sl).Log("msg", "Beginning probe", "command", module.Command, "timeout_seconds", timeoutSeconds)

	start := time.Now()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	//success := prober(ctx, target, module, registry, sl)
	sucess := false
	duration := time.Since(start).Seconds()
	probeDurationGauge.Set(duration)
	if success {
		probeSuccessGauge.Set(1)
		level.Info(sl).Log("msg", "Probe succeeded", "duration_seconds", duration)
	} else {
		level.Error(sl).Log("msg", "Probe failed", "duration_seconds", duration)
	}

	if r.URL.Query().Get("debug") == "true" {
		buf := &bytes.Buffer{}

		debugModule(buf, moduleName, &module)
		buf.Write('\n')
		debugRegistry(buf, registry)
		buf.Write('\n')
		debugLogger(buf, sl)

		w.Header().Set("Content-Type", "text/plain")
		w.Write(buf.Bytes())
		return
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func getTimeout(r *http.Request, baseTimeout time.Duration) (timeoutSeconds float64, err error) {
	// If a timeout is configured via the Prometheus header, add it to the request.
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		var err error
		timeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
	}
	if timeoutSeconds == 0 {
		timeoutSeconds = 120
	}

	var maxTimeoutSeconds = timeoutSeconds
	if baseTimeout.Seconds() < maxTimeoutSeconds && baseTimeout.Seconds() > 0 || maxTimeoutSeconds < 0 {
		timeoutSeconds = baseTimeout.Seconds()
	} else {
		timeoutSeconds = maxTimeoutSeconds
	}

	return timeoutSeconds, nil
}

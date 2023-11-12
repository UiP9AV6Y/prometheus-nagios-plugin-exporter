package prober

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/config"
	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/prober/nagios"
	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/template"
)

type Handler struct {
	namespace     string
	cache         *template.TemplateCache
	logger        log.Logger
	logLevel      level.Option
	debug         bool
	timeoutOffset float64
}

func NewHandler(namespace string, cache *template.TemplateCache, logger log.Logger, logLevel level.Option, debug bool, timeoutOffset float64) *Handler {
	result := &Handler{
		namespace:     namespace,
		logger:        logger,
		logLevel:      logLevel,
		cache:         cache,
		debug:         debug,
		timeoutOffset: timeoutOffset,
	}

	return result
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	maxTimeoutSeconds, err := h.getMaxTimeoutSeconds(r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse timeout from Prometheus header: %s", err), http.StatusBadRequest)
		return
	}

	moduleName, module, ok := config.FromContext(r.Context())
	if !ok {
		http.Error(w, "Unable to load module", http.StatusInternalServerError)
		return
	}

	data := nagios.NewLazyPluginBuilderContext(module.Variables, module.Environment).VisitVariables(nagios.MapVarsProvider(r.URL.Query())).VisitEnvironment(os.Getenv)

	metrics := nagios.NewPluginMetrics(module, h.namespace)
	builder := nagios.NewPluginBuilder(h.cache)
	prober, err := builder.Build(module, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create module probe %q", moduleName), http.StatusInternalServerError)
		level.Error(h.logger).Log("msg", "Unable to create module probe", "module", moduleName, "err", err)
		return
	}

	registry := prometheus.NewRegistry()
	if err := metrics.Register(registry); err != nil {
		http.Error(w, "Metrics setup failed", http.StatusInternalServerError)
		level.Error(h.logger).Log("msg", "Metrics setup failed", "err", err)
		return
	}

	timeoutSeconds := getTimeout(maxTimeoutSeconds, time.Duration(module.Timeout))
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds*float64(time.Second)))
	defer cancel()
	r = r.WithContext(ctx)

	logger := newScrapeLogger(h.logger, moduleName, h.logLevel)
	level.Info(logger).Log("msg", "Beginning probe", "command", prober.String(), "timeout_seconds", timeoutSeconds)

	start := time.Now()
	output, err := prober.Run(ctx)
	duration := time.Since(start).Seconds()

	if err != nil {
		level.Error(logger).Log("msg", "Probe execution failed", "duration_seconds", duration, "err", err)
	} else {
		if output.Error != nil {
			level.Warn(logger).Log("msg", "Probe evaluation failed", "duration_seconds", duration, "err", output.Error)
		} else {
			level.Info(logger).Log("msg", "Probe succeeded", "duration_seconds", duration)
		}

		level.Debug(logger).Log("msg", output.Output, "nagios_result", output.Status)
	}

	metrics.Report(output, err, duration)

	if debug, _ := strconv.ParseBool(r.URL.Query().Get("debug")); debug {
		if !h.debug {
			http.Error(w, "Debug feature has been disabled", http.StatusForbidden)
			level.Info(h.logger).Log("msg", "Rejecting request debugging")
			return
		}

		buf := &bytes.Buffer{}

		debugModule(buf, moduleName, module)
		buf.WriteByte('\n')
		debugPlugin(buf, prober, output, err)
		buf.WriteByte('\n')
		debugRegistry(buf, registry)
		buf.WriteByte('\n')
		debugLogger(buf, logger)

		w.Header().Set("Content-Type", "text/plain")
		w.Write(buf.Bytes())
		return
	}

	p := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	p.ServeHTTP(w, r)
}

func (h *Handler) getMaxTimeoutSeconds(v string) (maxTimeoutSeconds float64, err error) {
	if v != "" {
		if maxTimeoutSeconds, err = strconv.ParseFloat(v, 64); err != nil {
			return
		}
	}

	if maxTimeoutSeconds <= 0 {
		maxTimeoutSeconds = 120
	}

	maxTimeoutSeconds = maxTimeoutSeconds - h.timeoutOffset

	return
}

func getTimeout(maxTimeoutSeconds float64, baseTimeout time.Duration) (timeoutSeconds float64) {
	if baseTimeout.Seconds() < maxTimeoutSeconds && baseTimeout.Seconds() > 0 || maxTimeoutSeconds < 0 {
		timeoutSeconds = baseTimeout.Seconds()
	} else {
		timeoutSeconds = maxTimeoutSeconds
	}

	return timeoutSeconds
}

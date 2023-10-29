package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/config"
	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/prober"
)

const (
	title = "Nagios Plugin Exporter"
	ident = "nagios_plugin"
	name  = ident + "_exporter"

	healthEndpoint    = "/-/healthy"
	reloadEndpoint    = "/-/reload"
	telemetryEndpoint = "/metrics"
	probeEndpoint     = "/probe"
)

var (
	sc = config.NewSafeConfig(prometheus.DefaultRegisterer)

	configFile  = kingpin.Flag("config.file", "Nagios Plugin exporter configuration file.").Default(ident + ".yml").String()
	configCheck = kingpin.Flag("config.check", "If true validate the config file and then exit.").Default().Bool()

	logLevelProber = kingpin.Flag("log.prober", "Log level from probe requests. One of: [debug, info, warn, error, none]").Default("none").String()
	toolkitFlags   = webflag.AddFlags(kingpin.CommandLine, ":9665")

	moduleUnknownCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: ident + "_module_unknown_total",
		Help: "Count of unknown modules requested by probes",
	})
)

func init() {
	prometheus.MustRegister(version.NewCollector(name))
}

func main() {
	os.Exit(run())
}

func parseArgs() log.Logger {
	promlogConfig := &promlog.Config{}

	kingpin.CommandLine.UsageWriter(os.Stdout)
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print(name))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	return promlog.New(promlogConfig)
}

func watchConfig(reloadCh chan chan error, logger log.Logger) {
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)

	go func() {
		for {
			select {
			case <-hup:
				if err := sc.ReloadConfig(*configFile, logger); err != nil {
					level.Error(logger).Log("msg", "Error reloading config", "err", err)
					continue
				}
				level.Info(logger).Log("msg", "Reloaded config file")
			case rc := <-reloadCh:
				if err := sc.ReloadConfig(*configFile, logger); err != nil {
					level.Error(logger).Log("msg", "Error reloading config", "err", err)
					rc <- err
				} else {
					level.Info(logger).Log("msg", "Reloaded config file")
					rc <- nil
				}
			}
		}
	}()
}

func rootHandler() http.Handler {
	landingConfig := web.LandingConfig{
		Name:        title,
		Description: "Prometheus Exporter for Nagios Plugins",
		Version:     version.Info(),
		Form: web.LandingForm{
			Action: probeEndpoint,
			Inputs: []web.LandingFormInput{
				{
					Label:       "Target",
					Type:        "text",
					Name:        "target",
					Placeholder: "X.X.X.X/[::X]",
					Value:       "::1",
				},
				{
					Label:       "Module",
					Type:        "text",
					Name:        "module",
					Placeholder: "module",
					Value:       "http",
				},
			},
		},
		Links: []web.LandingLinks{
			{
				Address: telemetryEndpoint,
				Text:    "Metrics",
			},
		},
	}

	return web.NewLandingPage(landingConfig)
}

func healthHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	}
}

func probeHandlerFunc(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sc.Lock()
		conf := sc.C
		sc.Unlock()
		prober.Handler(w, r, conf, logger, moduleUnknownCounter, logLevelProber)
	}
}

func reloadHandlerFunc(reloadCh chan chan error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			rc := make(chan error)
			reloadCh <- rc
			if err := <-rc; err != nil {
				http.Error(w, fmt.Sprintf("failed to reload config: %s", err), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "POST method expected", http.StatusBadRequest)
		}
	}
}

func runServer(srvc chan struct{}, logger log.Logger) {
	srv := &http.Server{}

	go func() {
		if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
			level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
			close(srvc)
		}
	}()
}

func stopServer(srvc chan struct{}, logger log.Logger) int {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
			return 0
		case <-srvc:
			return 1
		}
	}
}

func run() int {
	logger := parseArgs()
	logLevelProberValue, _ := level.Parse(*logLevelProber)
	logLevelProber := level.Allow(logLevelProberValue)

	level.Info(logger).Log("msg", "Starting "+name, "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	if err := sc.ReloadConfig(*configFile, logger); err != nil {
		level.Error(logger).Log("msg", "Error loading config", "err", err)
		return 1
	}

	if *configCheck {
		level.Info(logger).Log("msg", "Config file is ok exiting...")
		return 0
	}

	level.Info(logger).Log("msg", "Loaded config file")

	reloadCh := make(chan chan error)
	watchConfig(reloadCh, logger)

	if landingPage, err := rootHandler(); err != nil {
		level.Error(logger).Log("msg", "Unable to set up root handler", "err", err)
		return 1
	} else {
		http.Handle("/", landingPage)
	}

	http.Handle(telemetryEndpoint, promhttp.Handler())
	http.HandleFunc(reloadEndpoint, reloadHandlerFunc(reloadCh))
	http.HandleFunc(healthEndpoint, healthHandlerFunc())
	http.HandleFunc(probeEndpoint, probeHandlerFunc())

	srvc := make(chan struct{})
	runServer(srvc, logger)
	stopServer(srvc, logger)

}

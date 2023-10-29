package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Config struct {
	Modules map[string]Module `yaml:"modules,omitempty"`
}

type SafeConfig struct {
	sync.RWMutex
	C                   *Config
	configReloadSuccess prometheus.Gauge
	configReloadSeconds prometheus.Gauge
}

func NewSafeConfig(namespace string, reg prometheus.Registerer) *SafeConfig {
	configReloadSuccess := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "config_last_reload_successful",
		Help:      "Exporter config loaded successfully.",
	})
	configReloadSeconds := promauto.With(reg).NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "config_last_reload_success_timestamp_seconds",
		Help:      "Timestamp of the last successful configuration reload.",
	})
	config := &Config{}
	result := &SafeConfig{
		C:                   config,
		configReloadSuccess: configReloadSuccess,
		configReloadSeconds: configReloadSeconds,
	}

	return result
}

func (sc *SafeConfig) ReloadConfig(confFile string, logger log.Logger) (err error) {
	var c = &Config{}
	defer func() {
		if err != nil {
			sc.configReloadSuccess.Set(0)
		} else {
			sc.configReloadSuccess.Set(1)
			sc.configReloadSeconds.SetToCurrentTime()
		}
	}()

	yamlReader, err := os.Open(confFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)

	if err = decoder.Decode(c); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	for name, module := range c.Modules {
		if module.HTTP.NoFollowRedirects != nil {
			// Hide the old flag from the /config page.
			module.HTTP.NoFollowRedirects = nil
			c.Modules[name] = module
			if logger != nil {
				level.Warn(logger).Log("msg", "no_follow_redirects is deprecated and will be removed in the next release. It is replaced by follow_redirects.", "module", name)
			}
		}
	}

	sc.UpdateConfig(c)

	return nil
}

func (sc *SafeConfig) UpdateConfig(c *Config) {
	sc.Lock()
	sc.C = c
	sc.Unlock()
}

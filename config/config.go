package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Config defines the configuration root node
type Config struct {
	Modules map[string]Module `yaml:"modules,omitempty"`
}

// YAML renders the instance as YAML representation
func (c *Config) MarshalYAML() ([]byte, error) {
	type rawConfig Config
	return yaml.Marshal((*rawConfig)(c))
}

// SafeConfig is thread-safe Config instance provider
type SafeConfig struct {
	sync.RWMutex
	config              *Config
	configReloadSuccess prometheus.Gauge
	configReloadSeconds prometheus.Gauge
}

// NewSafeConfig creates a new SafeConfig instance
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
		config:              config,
		configReloadSuccess: configReloadSuccess,
		configReloadSeconds: configReloadSeconds,
	}

	return result
}

// ReloadConfig reads the configuration from the given path and updates its
// internal config instance with the parsed result
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

	r, err := os.Open(confFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}
	defer r.Close()
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)

	if err = decoder.Decode(c); err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	sc.UpdateConfig(c)

	return nil
}

// UpdateConfig replaces the internal config instance with the given one (thread-safe)
func (sc *SafeConfig) UpdateConfig(c *Config) {
	sc.Lock()
	sc.config = c
	sc.Unlock()
}

// ProvideConfig is a thread-safe visitor implementation, to gain access to the
// internal config instance.
func (sc *SafeConfig) ProvideConfig(visitor func(*Config)) {
	sc.Lock()
	conf := sc.config
	sc.Unlock()

	visitor(conf)
}

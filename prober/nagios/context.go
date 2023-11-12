package nagios

import (
	"os"

	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/config"
	monitoring "github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/nagios"
)

// PluginBuilderContext is a template execution context
type PluginBuilderContext struct {
	Vars map[string][]string
	Env  map[string]string
}

// MapVarsProvider is a wrapper around the map lookup operator,
// but instead of returning 2 variables, the result is simply nil
// if the lookup fails
func MapVarsProvider(m map[string][]string) func(string) []string {
	return func(s string) []string {
		if v, ok := m[s]; ok {
			return v
		}

		return nil
	}
}

// NewLazyPluginBuilderContext creates a new plugin builder context. the variables are
// copied as is to match the type signature of the internal struct member.
func NewLazyPluginBuilderContext(vars map[string]config.LazyArray, env map[string]string) *PluginBuilderContext {
	vs := make(map[string][]string, len(vars))
	for k, v := range vars {
		vs[k] = []string(v)
	}

	return NewPluginBuilderContext(vs, env)
}

// NewPluginBuilderContext creates a new plugin builder context
func NewPluginBuilderContext(vars map[string][]string, env map[string]string) *PluginBuilderContext {
	result := &PluginBuilderContext{
		Vars: vars,
		Env:  env,
	}

	return result
}

// VisitVariables populates the instance Vars with values from the provider.
// If the provider does not yield a value for the map keys, the original value
// is used as fallback. If the result of that operation yields an empty value
// (i.e. nil or an empty slice), the key will be remove from the instance map.
//
// Empty values are removed from the map values as well
func (c *PluginBuilderContext) VisitVariables(provider func(string) []string) *PluginBuilderContext {
	result := make(map[string][]string, len(c.Vars))
	for k, v := range c.Vars {
		d := provider(k)
		if d == nil || len(d) == 0 {
			d = v
		}

		d = monitoring.Compact(d)
		if d != nil && len(d) > 0 {
			result[k] = d
		}
	}

	c.Vars = result

	return c
}

// VisitEnvironment populates the instance Vars with values from the provider.
// If the provider does not yield a value for the map keys, the original value
// is used as fallback. If the result of that operation yields an empty value
// (i.e. an empty string), the key will be remove from the instance map.
func (c *PluginBuilderContext) VisitEnvironment(provider func(string) string) *PluginBuilderContext {
	result := make(map[string]string, len(c.Env))
	for k, v := range c.Env {
		d := provider(k)
		if d == "" {
			d = os.Expand(v, provider)
		}
		if d != "" {
			result[k] = d
		}
	}

	c.Env = result

	return c
}

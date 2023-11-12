package nagios

import (
	"cmp"
	"errors"
	"slices"
	"strings"

	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/config"
	monitoring "github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/nagios"
	"github.com/UiP9AV6Y/prometheus-nagios-plugin-exporter/template"
)

var errMissingCommand = errors.New("Module is missing the plugin command")

type PluginBuilder struct {
	cache *template.TemplateCache
}

// argument is the rendered version of config.Argument
type argument struct {
	condition bool
	value     []string
	order     int
	key       string
	required  bool
	repeatKey bool
	skipKey   bool
	separator string
}

func NewPluginBuilder(cache *template.TemplateCache) *PluginBuilder {
	result := &PluginBuilder{
		cache: cache,
	}

	return result
}

// JoinKeyValues create a slice of the provided map,
// joining the keys with their values using the defined separator
func JoinKeyValues(m map[string]string, sep string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, k+sep+v)
	}

	return result
}

func (b *PluginBuilder) Build(module *config.Module, ctx *PluginBuilderContext) (*monitoring.Plugin, error) {
	args, err := b.parseArguments(module.Arguments, ctx)
	if err != nil {
		return nil, err
	}

	if module.Command == "" {
		return nil, errMissingCommand
	}

	result := monitoring.NewPlugin(module.Command, args, JoinKeyValues(ctx.Env, "="))

	return result, nil
}

func renderArguments(argv []*argument) []string {
	result := make([]string, 0, len(argv))

	slices.SortFunc(argv, func(a, b *argument) int {
		return cmp.Compare(a.order, b.order)
	})

	for _, arg := range argv {
		if !arg.condition || arg.key == "" {
			continue
		}

		if len(arg.value) == 0 && !arg.skipKey {
			result = append(result, arg.key)
			continue
		}

		first := true
		for _, val := range arg.value {
			if arg.skipKey || (!first && !arg.repeatKey) {
				if val == "" {
					continue
				}

				result = append(result, val)
			} else if val == "" {
				result = append(result, arg.key)
			} else if arg.separator == " " {
				result = append(result, arg.key, val)
			} else {
				result = append(result, arg.key+arg.separator+val)
			}

			first = false
		}
	}

	return result
}

func (b *PluginBuilder) parseArguments(args map[string]config.Argument, ctx *PluginBuilderContext) ([]string, error) {
	argv := make([]*argument, 0, len(args))

	for key, arg := range args {
		item, err := b.parseArgument(&arg, ctx)
		if err != nil {
			return nil, err
		}

		if item.key == "" {
			item.key = key
		}

		argv = append(argv, item)
	}

	return renderArguments(argv), nil
}

func (b *PluginBuilder) parseArgument(c *config.Argument, ctx *PluginBuilderContext) (*argument, error) {
	var result = &argument{}
	var err error

	result.order = c.Order
	result.key = c.Key
	result.separator = c.Separator

	result.value = make([]string, 0, len(c.Value))
	for _, v := range c.Value {
		val, err := b.cache.RenderString("Value", v, ctx)
		if err != nil {
			return nil, err
		}

		for _, v2 := range strings.Split(val, "\n") {
			v2 = strings.TrimSpace(v2)
			if v2 != "" {
				result.value = append(result.value, v2)
			}
		}
	}

	if c.Condition != "" {
		result.condition, err = b.cache.RenderBool("Condition", string(c.Condition), ctx)
		// omit rendering the rest if the argument is not used anyway
		if err != nil || !result.condition {
			return result, err
		}
	} else {
		result.condition = len(result.value) > 0
	}

	if c.Required != "" {
		result.required, err = b.cache.RenderBool("Required", string(c.Required), ctx)
		if err != nil {
			return nil, err
		}
	}

	if c.RepeatKey != "" {
		result.repeatKey, err = b.cache.RenderBool("RepeatKey", string(c.RepeatKey), ctx)
		if err != nil {
			return nil, err
		}
	} else {
		result.repeatKey = true
	}

	if c.SkipKey != "" {
		result.skipKey, err = b.cache.RenderBool("SkipKey", string(c.SkipKey), ctx)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

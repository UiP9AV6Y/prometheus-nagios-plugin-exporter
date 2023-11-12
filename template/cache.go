package template

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

const TemplateToken = "{{"

var Functions template.FuncMap

func init() {
	s := sprig.GenericFuncMap()
	Functions = template.FuncMap{
		"lines":     Lines,
		"net_host":  NetHost,
		"net_port":  NetPort,
		"read_file": ReadFile,
		"lower":     strings.ToLower,
		"trim":      strings.TrimSpace,
		"upper":     strings.ToUpper,
		"compact":   s["mustCompact"],
		"default":   s["default"],
		"first":     s["mustFirst"],
		"initial":   s["mustInitial"],
		"join":      s["join"],
		"rest":      s["mustLast"],
		"strval":    s["toString"],
		"uniq":      s["mustUniq"],
	}
}

type TemplateCache struct {
	sync.RWMutex

	cache   map[string]*template.Template
	factory func(string) *template.Template
}

func NewFuncMapTemplateCache(m template.FuncMap) *TemplateCache {
	return NewTemplateCache(func(name string) *template.Template {
		return template.New(name).Funcs(m).Option("missingkey=zero")
	})
}

func NewTemplateCache(factory func(string) *template.Template) *TemplateCache {
	cache := make(map[string]*template.Template)
	result := &TemplateCache{
		cache:   cache,
		factory: factory,
	}

	return result
}

func (c *TemplateCache) Flush() {
	c.Lock()
	c.cache = make(map[string]*template.Template)
	c.Unlock()
}

func (c *TemplateCache) RenderString(i, s string, ctx interface{}) (string, error) {
	if s == "" || strings.Index(s, TemplateToken) == -1 {
		return s, nil
	}

	tmpl, err := c.get(i, s)
	if err != nil {
		return "", nil
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, ctx); err != nil {
		return "", err
	}

	return result.String(), nil
}

func (c *TemplateCache) RenderBool(i, b string, ctx interface{}) (bool, error) {
	s, err := c.RenderString(i, b, ctx)
	if err != nil {
		return false, err
	} else if s == "" {
		return false, nil
	}

	return strconv.ParseBool(s)
}

func (c *TemplateCache) get(i, s string) (tmpl *template.Template, err error) {
	var ok bool
	c.Lock()
	tmpl, ok = c.cache[s]
	c.Unlock()

	if ok {
		return tmpl, nil
	}

	tmpl, err = c.factory(i).Parse(s)
	if err != nil {
		return
	}

	c.Lock()
	c.cache[s] = tmpl
	c.Unlock()

	return
}

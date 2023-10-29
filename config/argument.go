package config

type Argument struct {
	Condition string `yaml:"set_if,omitempty"`
	Value     string `yaml:"value,omitempty"`
	Order     int    `yaml:"order,omitempty"`
	Key       string `yaml:"key,omitempty"`
	Required  bool   `yaml:"required,omitempty"`
	RepeatKey bool   `yaml:"repeat_key,omitempty"`
	SkipKey   bool   `yaml:"skip_key,omitempty"`
	Separator string `yaml:"separator,omitempty"`
}

func (a *Argument) MarshalIcinga(name string) ([]byte, error) {
	result := name + ` = "` + a.Value + `"`

	return []byte(result)
}

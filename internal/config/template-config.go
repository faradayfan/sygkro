package config

var (
	TemplateConfigFileName = "sygkro.template.yaml"
)

type TemplateConfig struct {
	Name        string           `yaml:"name"`
	Description string           `yaml:"description"`
	Version     string           `yaml:"version"`
	Templating  TemplatingConfig `yaml:"templating"`
}

type TemplatingConfig struct {
	Inputs map[string]string `yaml:"inputs"`
}

func (s *TemplateConfig) Write(path string) error {
	return WriteYAML(path, s)
}

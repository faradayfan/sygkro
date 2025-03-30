package config

const (
	SyncConfigFileName = ".sygkro.sync.yaml"
)

type SyncConfig struct {
	Source SourceConfig      `yaml:"source"`
	Inputs map[string]string `yaml:"inputs"`
}

type SourceConfig struct {
	TemplatePath        string `yaml:"template_path"`
	TemplateName        string `yaml:"template_name"`
	TemplateVersion     string `yaml:"template_version"`
	TemplateTrackingRef string `yaml:"template_tracking_ref"`
}

func (s *SyncConfig) Write(path string) error {
	return WriteYAML(path, s)
}

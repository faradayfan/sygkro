package config

const (
	SyncConfigFileName = ".sygkro.sync.yaml"
)

type SyncConfig struct {
	Path   string            `yaml:"-"` // ignore when serializing
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

func ReadSyncConfig(path string) (*SyncConfig, error) {
	syncConfig := &SyncConfig{}
	err := ReadYAML(path, syncConfig)

	syncConfig.Path = path

	if err != nil {
		return nil, err
	}

	return syncConfig, nil
}

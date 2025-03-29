package templates

import "embed"

//go:embed *.tpl
var FS embed.FS

func GetTemplate(name string) ([]byte, error) {
	tpl, err := FS.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

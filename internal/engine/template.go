package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/faradayfan/sygkro/internal/config"
)

func RenderString(tmplStr string, data map[string]string) (string, error) {
	tmpl, err := template.New("render").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return buf.String(), nil
}

func ProcessTemplateDir(sourceDir, targetDir string, inputs map[string]string, opts *config.TemplateOptions) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		renderedRelPath, err := RenderString(relPath, inputs)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, renderedRelPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		if opts != nil {
			for _, pattern := range opts.SkipRender {
				if matched, _ := filepath.Match(pattern, relPath); matched {
					// Copy the file without rendering.
					content, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					return os.WriteFile(targetPath, content, info.Mode())
				}
			}
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		processed, rawMap, err := PreprocessRawBlocks(string(content))

		if err != nil {
			return err
		}

		rendered, err := RenderString(processed, inputs)
		if err != nil {
			return err
		}

		finalOutput := PostprocessRawBlocks(rendered, rawMap)

		return os.WriteFile(targetPath, []byte(finalOutput), info.Mode())
	})
}

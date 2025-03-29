package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/faradayfan/sygkro/internal/config"
)

// RenderString renders a template string using the provided data.
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

// ProcessTemplateDir recursively walks through the source directory,
// rendering file paths and file contents using the provided inputs,
// and writes the output to the target directory.
func ProcessTemplateDir(sourceDir, targetDir string, inputs map[string]string, opts *config.TemplateOptions) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Compute the relative path from the source directory.
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Render the relative path using the provided inputs.
		renderedRelPath, err := RenderString(relPath, inputs)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, renderedRelPath)

		if info.IsDir() {
			// Create the target directory.
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

		// Read file content.
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Preprocess the content if needed.
		processed, rawMap, err := PreprocessRawBlocks(string(content))

		// Render the file content.
		rendered, err := RenderString(processed, inputs)
		if err != nil {
			return err
		}

		finalOutput := PostprocessRawBlocks(rendered, rawMap)

		// Write the rendered content to the target path.
		return os.WriteFile(targetPath, []byte(finalOutput), info.Mode())
	})
}

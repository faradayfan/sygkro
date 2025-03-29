package engine

import (
	"fmt"
	"regexp"
	"strings"
)

// PreprocessRawBlocks scans the input content for raw blocks marked by
// "{{/* no_render:start */}}" and "{{/* no_render:end */}}". It replaces each
// raw block with a unique placeholder and returns the processed content along with
// a map of placeholders to their original content.
func PreprocessRawBlocks(content string) (string, map[string]string, error) {
	// Regex explanation:
	// (?s)                   : Enable dot-all mode so '.' matches newline.
	// {{/\*\s*no_render:start\s*\*/}} : Matches the start marker.
	// (.*?)                 : Lazily captures everything until the end marker.
	// {{/\*\s*no_render:end\s*\*/}}   : Matches the end marker.
	pattern := `(?s){{/\*\s*no_render:start\s*\*/}}(.*?){{/\*\s*no_render:end\s*\*/}}`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", nil, fmt.Errorf("failed to compile regex: %w", err)
	}

	rawBlocks := make(map[string]string)
	placeholderIndex := 0

	// Replace each found raw block with a unique placeholder.
	processedContent := re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the inner content.
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 2 {
			// If no match group, return the original content.
			return match
		}
		rawContent := submatches[1]
		placeholder := fmt.Sprintf("__NO_RENDER_BLOCK_%d__", placeholderIndex)
		rawBlocks[placeholder] = rawContent
		placeholderIndex++
		return placeholder
	})

	return processedContent, rawBlocks, nil
}

// PostprocessRawBlocks replaces placeholders in the rendered content with their
// corresponding raw block content.
func PostprocessRawBlocks(content string, rawBlocks map[string]string) string {
	for placeholder, rawContent := range rawBlocks {
		content = strings.ReplaceAll(content, placeholder, rawContent)
	}
	return content
}

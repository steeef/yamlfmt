// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package basic

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
)

const PreserveBackslashFormatterType string = "preserve_backslash"

// PreserveBackslashFormatter extends BasicFormatter to preserve backslash-continued strings
type PreserveBackslashFormatter struct {
	*BasicFormatter
}

// Regex to detect backslash-continued strings in YAML values
var backslashContinuationRegex = regexp.MustCompile(`(?m)^(\s*[^:\n]+:\s*)"([^"\n]*\\\s*)$`)

// BackslashPreservation holds info about a preserved string
type BackslashPreservation struct {
	Placeholder string
	Original    string
	Indent      string
}

func (f *PreserveBackslashFormatter) Type() string {
	return PreserveBackslashFormatterType
}

func (f *PreserveBackslashFormatter) Format(input []byte) ([]byte, error) {
	// Text-level approach: detect and replace backslash continuations with placeholders
	modifiedInput, preservations, err := f.preprocessBackslashContinuations(input)
	if err != nil {
		return nil, err
	}

	// Apply standard basic formatting to the modified input
	formatted, err := f.BasicFormatter.Format(modifiedInput)
	if err != nil {
		return nil, err
	}

	// Post-process: restore original backslash continuations
	result := f.postprocessBackslashContinuations(formatted, preservations)
	return result, nil
}

// preprocessBackslashContinuations detects backslash-continued strings and replaces them with placeholders
func (f *PreserveBackslashFormatter) preprocessBackslashContinuations(content []byte) ([]byte, []BackslashPreservation, error) {
	input := string(content)
	var preservations []BackslashPreservation

	// Find all backslash-continued multi-line strings
	lines := strings.Split(input, "\n")
	modifiedLines := make([]string, len(lines))
	copy(modifiedLines, lines)

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Look for lines that start a backslash-continued string
		if match := f.findBackslashContinuationStart(line); match != nil {
			// Extract the complete multi-line string
			originalString, endLine := f.extractBackslashContinuedString(lines, i)

			// Generate unique placeholder
			placeholder := f.generatePlaceholder()

			// Store preservation info
			preservations = append(preservations, BackslashPreservation{
				Placeholder: placeholder,
				Original:    originalString,
				Indent:      match.indent,
			})

			// Replace the multi-line string with placeholder in modified content
			modifiedLines[i] = match.prefix + `"` + placeholder + `"`

			// Clear the continuation lines
			for j := i + 1; j <= endLine; j++ {
				modifiedLines[j] = ""
			}

			// Skip to after the end of this multi-line string
			i = endLine
		}
	}

	// Remove empty lines that were cleared
	var finalLines []string
	for _, line := range modifiedLines {
		if line != "" || len(finalLines) == 0 || finalLines[len(finalLines)-1] != "" {
			finalLines = append(finalLines, line)
		}
	}

	return []byte(strings.Join(finalLines, "\n")), preservations, nil
}

// BackslashMatch represents a matched backslash continuation pattern
type BackslashMatch struct {
	prefix string
	indent string
}

// findBackslashContinuationStart checks if a line starts a backslash-continued string
func (f *PreserveBackslashFormatter) findBackslashContinuationStart(line string) *BackslashMatch {
	// Pattern: key: "value\   (with potential whitespace after backslash)
	pattern := regexp.MustCompile(`^(\s*)([^:]+:\s*)"([^"]*\\\s*)$`)
	matches := pattern.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil
	}
	return &BackslashMatch{
		prefix: matches[1] + matches[2], // indent + key + colon + space
		indent: matches[1],              // just the indent
	}
}

// extractBackslashContinuedString extracts the complete multi-line backslash-continued string
func (f *PreserveBackslashFormatter) extractBackslashContinuedString(lines []string, startLine int) (string, int) {
	var stringLines []string
	stringLines = append(stringLines, lines[startLine])

	endLine := startLine
	for i := startLine + 1; i < len(lines); i++ {
		line := lines[i]
		stringLines = append(stringLines, line)
		endLine = i

		// Check if this line ends the string (ends with quote, not backslash-quote)
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, `"`) && !strings.HasSuffix(trimmed, `\"`) {
			break
		}
	}

	return strings.Join(stringLines, "\n"), endLine
}

// generatePlaceholder creates a unique placeholder token
func (f *PreserveBackslashFormatter) generatePlaceholder() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("__YAMLFMT_BACKSLASH_%x__", b)
}

// postprocessBackslashContinuations restores original backslash-continued strings from placeholders
func (f *PreserveBackslashFormatter) postprocessBackslashContinuations(formatted []byte, preservations []BackslashPreservation) []byte {
	result := string(formatted)

	// Replace each placeholder with its original backslash-continued string
	for _, preservation := range preservations {
		// Find the placeholder in the formatted output
		placeholderPattern := `"` + regexp.QuoteMeta(preservation.Placeholder) + `"`
		pattern := regexp.MustCompile(placeholderPattern)

		// Replace with the original multi-line string
		result = pattern.ReplaceAllString(result, preservation.Original)
	}

	return []byte(result)
}

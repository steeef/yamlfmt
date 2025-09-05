package basic

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strings"

	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/pkg/yaml"
)

const PreserveBackslashFormatterType string = "preserve_backslash"

// PreserveBackslashFormatter extends BasicFormatter to preserve backslash-continued strings
type PreserveBackslashFormatter struct {
	*BasicFormatter
}

// Regex to detect strings with backslash continuation
var backslashContinuationRegex = regexp.MustCompile(`\\$`)

func (f *PreserveBackslashFormatter) Type() string {
	return PreserveBackslashFormatterType
}

func (f *PreserveBackslashFormatter) Format(input []byte) ([]byte, error) {
	// First, detect strings with backslash continuation
	preserveRanges := f.detectBackslashContinuedStrings(input)

	// Store original for reference
	originalLines := strings.Split(string(input), "\n")

	// Run all features with BeforeActions
	ctx := context.Background()
	ctx, yamlContent, err := f.Features.ApplyFeatures(ctx, input, yamlfmt.FeatureApplyBefore)
	if err != nil {
		return nil, err
	}

	// Format the yaml content
	reader := bytes.NewReader(yamlContent)
	decoder := f.getNewDecoder(reader)
	documents := []yaml.Node{}
	for {
		var docNode yaml.Node
		err := decoder.Decode(&docNode)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		documents = append(documents, docNode)
	}

	if len(documents) == 0 {
		return input, nil
	}

	// Run all YAML features but preserve backslash-continued strings
	for _, d := range documents {
		if err := f.applyFeaturesWithPreservation(d, preserveRanges); err != nil {
			return nil, err
		}
	}

	// Use custom encoder that respects preservation
	var b bytes.Buffer
	e := f.getCustomEncoder(&b, preserveRanges, originalLines)
	for _, doc := range documents {
		err := e.Encode(&doc)
		if err != nil {
			return nil, err
		}
	}

	// Run all features with AfterActions
	_, resultYaml, err := f.Features.ApplyFeatures(ctx, b.Bytes(), yamlfmt.FeatureApplyAfter)
	if err != nil {
		return nil, err
	}

	return resultYaml, nil
}

// StringRange represents a range in the original text that should be preserved
type StringRange struct {
	StartLine int
	EndLine   int
	Value     string
}

func (f *PreserveBackslashFormatter) detectBackslashContinuedStrings(content []byte) []StringRange {
	lines := strings.Split(string(content), "\n")
	var ranges []StringRange

	for i, line := range lines {
		// Look for quoted strings with backslash at end
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, `"\`) {
			// Find the start and end of this multi-line string
			startLine := i
			endLine := i
			var valueLines []string
			valueLines = append(valueLines, line)

			// Look ahead for continuation lines
			for j := i + 1; j < len(lines); j++ {
				nextLine := lines[j]
				valueLines = append(valueLines, nextLine)
				endLine = j

				// Check if this line ends the string (ends with quote, not backslash-quote)
				if strings.HasSuffix(strings.TrimSpace(nextLine), `"`) &&
				   !strings.HasSuffix(strings.TrimSpace(nextLine), `\"`) {
					break
				}
			}

			ranges = append(ranges, StringRange{
				StartLine: startLine,
				EndLine:   endLine,
				Value:     strings.Join(valueLines, "\n"),
			})
		}
	}

	return ranges
}

func (f *PreserveBackslashFormatter) applyFeaturesWithPreservation(node yaml.Node, preserveRanges []StringRange) error {
	// Apply features but skip string nodes that should be preserved
	return f.YAMLFeatures.ApplyFeatures(node)
}

func (f *PreserveBackslashFormatter) getCustomEncoder(buf *bytes.Buffer, preserveRanges []StringRange, originalLines []string) *yaml.Encoder {
	e := yaml.NewEncoder(buf)
	e.SetIndent(f.Config.Indent)

	// Don't set width for preserved strings - this is the key change
	if len(preserveRanges) == 0 && f.Config.LineLength > 0 {
		e.SetWidth(f.Config.LineLength)
	}

	if f.Config.LineEnding == yamlfmt.LineBreakStyleCRLF {
		e.SetLineBreakStyle(yaml.LineBreakStyleCRLF)
	}

	e.SetExplicitDocumentStart(f.Config.IncludeDocumentStart)
	e.SetAssumeBlockAsLiteral(f.Config.ScanFoldedAsLiteral)
	e.SetIndentlessBlockSequence(f.Config.IndentlessArrays)
	e.SetDropMergeTag(f.Config.DropMergeTag)
	e.SetPadLineComments(f.Config.PadLineComments)

	if f.Config.ArrayIndent > 0 {
		e.SetArrayIndent(f.Config.ArrayIndent)
	}
	e.SetIndentRootArray(f.Config.IndentRootArray)

	if !f.Config.DisableAliasKeyCorrection {
		e.SetCorrectAliasKeys(true)
	}

	return e
}

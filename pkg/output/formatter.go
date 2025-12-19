package output

import (
	"fmt"
	"io"

	"github.com/aboigues/k8t/pkg/types"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	FormatTypeText OutputFormat = "text"
	FormatTypeJSON OutputFormat = "json"
	FormatTypeYAML OutputFormat = "yaml"
)

// Format writes the analysis report in the specified format
func Format(report *types.AnalysisReport, format OutputFormat, noColor bool, w io.Writer) error {
	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}

	switch format {
	case FormatTypeText:
		return formatTextOutput(report, noColor, w)
	case FormatTypeJSON:
		return formatJSONOutput(report, w)
	case FormatTypeYAML:
		return formatYAMLOutput(report, w)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// ParseFormat converts a string to OutputFormat
func ParseFormat(s string) (OutputFormat, error) {
	switch s {
	case "text", "":
		return FormatTypeText, nil
	case "json":
		return FormatTypeJSON, nil
	case "yaml", "yml":
		return FormatTypeYAML, nil
	default:
		return "", fmt.Errorf("unsupported format '%s': must be one of: text, json, yaml", s)
	}
}

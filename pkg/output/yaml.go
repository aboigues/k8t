package output

import (
	"io"

	"github.com/aboigues/k8t/pkg/types"
	"gopkg.in/yaml.v3"
)

// formatYAMLOutput renders report as YAML
func formatYAMLOutput(report *types.AnalysisReport, w io.Writer) error {
	if report == nil {
		return nil
	}

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	defer encoder.Close()

	return encoder.Encode(report)
}

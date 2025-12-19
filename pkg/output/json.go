package output

import (
	"encoding/json"
	"io"

	"github.com/aboigues/k8t/pkg/types"
)

// formatJSONOutput renders report as pretty-printed JSON
func formatJSONOutput(report *types.AnalysisReport, w io.Writer) error {
	if report == nil {
		return nil
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

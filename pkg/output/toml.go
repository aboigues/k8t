package output

import (
	"io"

	"github.com/aboigues/k8t/pkg/types"
	"github.com/pelletier/go-toml/v2"
)

// formatTOMLOutput renders report as TOML
func formatTOMLOutput(report *types.AnalysisReport, w io.Writer) error {
	if report == nil {
		return nil
	}

	encoder := toml.NewEncoder(w)
	encoder.SetIndentTables(true)
	return encoder.Encode(report)
}

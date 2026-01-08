package output

import (
	"encoding/xml"
	"io"

	"github.com/aboigues/k8t/pkg/types"
)

// formatXMLOutput renders report as pretty-printed XML
func formatXMLOutput(report *types.AnalysisReport, w io.Writer) error {
	if report == nil {
		return nil
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return err
	}

	// Write newline at the end for better formatting
	_, err := w.Write([]byte("\n"))
	return err
}

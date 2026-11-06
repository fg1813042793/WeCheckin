package workflowsummary

import "math"

const (
	exportBaseGridColumns = 24
	exportFineGridColumns = 120
)

func normalizedExportSpan(span float64) float64 {
	if span <= 0 || span > exportBaseGridColumns || math.IsNaN(span) {
		return exportBaseGridColumns
	}
	return span
}

func exportFieldGridColumns(rows [][]exportFieldCell) int {
	for _, row := range rows {
		for _, field := range row {
			if field.Span == 4.8 {
				return exportFineGridColumns
			}
		}
	}
	return exportBaseGridColumns
}

// The stored span always uses /24 semantics. Only integer document grids scale
// to 120 when needed; a fifth is 24 of 120, never rounded to 4 or 5 of 24.
func exportGridSpan(span float64, columns int) int {
	return int(math.Round(normalizedExportSpan(span) * float64(columns) / exportBaseGridColumns))
}

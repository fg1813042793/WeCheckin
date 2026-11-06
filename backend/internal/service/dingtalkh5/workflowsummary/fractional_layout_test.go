package workflowsummary

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"wecheckin/backend/internal/workflowcore"
)

func TestFractionalFormRowsPreserveWidthsAndBoundaries(t *testing.T) {
	for _, spans := range [][]float64{
		{4.8, 4.8, 4.8, 4.8, 4.8}, {4, 4, 4, 4, 4, 4},
		{3, 3, 3, 3, 3, 3, 3, 3}, {18, 3, 3}, {8, 8, 8},
	} {
		t.Run(fmt.Sprint(spans), func(t *testing.T) {
			fields := fractionalFormFields(t, spans)
			fields = append(fields, workflowcore.FormField{Key: "legacy", Label: "旧字段", Type: workflowcore.FormFieldTypeText})
			for _, grouped := range []bool{false, true} {
				form := fields
				if grouped {
					form = []workflowcore.FormField{{Key: "group", Label: "分组", Type: workflowcore.FormFieldTypeGroup, Fields: fields}}
				}
				sections := exportFormSections(form, nil, nil, true)
				if len(sections) != 1 || len(sections[0].FieldRows) != 2 {
					t.Fatalf("grouped=%v: expected full row followed by default-width row, got %#v", grouped, sections)
				}
				rows := sections[0].FieldRows
				if len(rows[0]) != len(spans) || len(rows[1]) != 1 || rows[1][0].Span != 24 {
					t.Fatalf("row boundaries changed: %#v", rows)
				}
				for i, want := range spans {
					if got := float64(rows[0][i].Span); got != want {
						t.Fatalf("span[%d] = %v, want %v", i, got, want)
					}
				}
				if fields[len(fields)-1].Span != 0 {
					t.Fatal("export mutated the legacy definition")
				}
			}
		})
	}
}

func TestFractionalExportColumnWidthsAndMerges(t *testing.T) {
	for _, tc := range []struct {
		name            string
		spans           []float64
		grid, gridWidth int
		widths, merges  []int
	}{
		{"fifths", []float64{4.8, 4.8, 4.8, 4.8, 4.8}, 120, 85, []int{2040, 2040, 2040, 2040, 2040}, []int{24, 24, 24, 24, 24}},
		{"eighths_mix", []float64{18, 3, 3}, 24, 425, []int{7650, 1275, 1275}, []int{18, 3, 3}},
		{"sixths", []float64{4, 4, 4, 4, 4, 4}, 24, 425, []int{1700, 1700, 1700, 1700, 1700, 1700}, []int{4, 4, 4, 4, 4, 4}},
		{"legacy_thirds", []float64{8, 8, 8}, 24, 425, []int{3400, 3400, 3400}, []int{8, 8, 8}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sections := exportFormSections(fractionalFormFields(t, tc.spans), nil, nil, true)
			document := exportDocument{Title: "宽度回归", Sections: sections}
			body, err := renderDOCX(document)
			if err != nil {
				t.Fatal(err)
			}
			var word struct {
				Tables []struct {
					Grid []struct {
						Width int `xml:"w,attr"`
					} `xml:"tblGrid>gridCol"`
					Rows []struct {
						Cells []struct {
							Properties struct {
								Width struct {
									Value int `xml:"w,attr"`
								} `xml:"tcW"`
								Span struct {
									Value int `xml:"val,attr"`
								} `xml:"gridSpan"`
							} `xml:"tcPr"`
						} `xml:"tc"`
					} `xml:"tr"`
				} `xml:"body>tbl"`
			}
			if err := xml.Unmarshal([]byte(readZipEntries(t, body)["word/document.xml"]), &word); err != nil {
				t.Fatal(err)
			}
			if len(word.Tables) != 2 || len(word.Tables[1].Grid) != tc.grid {
				t.Fatalf("unexpected Word grid: %#v", word)
			}
			for _, column := range word.Tables[1].Grid {
				if column.Width != tc.gridWidth {
					t.Fatalf("Word column width = %d, want %d", column.Width, tc.gridWidth)
				}
			}
			if len(word.Tables[1].Rows) != 2 {
				t.Fatalf("expected label/value rows")
			}
			for _, row := range word.Tables[1].Rows {
				var widths, merges []int
				for _, cell := range row.Cells {
					widths = append(widths, cell.Properties.Width.Value)
					merges = append(merges, cell.Properties.Span.Value)
				}
				if !reflect.DeepEqual(widths, tc.widths) || !reflect.DeepEqual(merges, tc.merges) {
					t.Fatalf("Word widths=%v merges=%v", widths, merges)
				}
			}

			body, err = renderXLSX([]exportDocument{document})
			if err != nil {
				t.Fatal(err)
			}
			worksheet := readZipEntries(t, body)["xl/worksheets/sheet1.xml"]
			columnXML := fmt.Sprintf(`<col min="1" max="%d" width="%g" customWidth="1"/>`, tc.grid, 4.5*24/float64(tc.grid))
			if !strings.Contains(worksheet, columnXML) {
				t.Fatalf("XLSX missing column widths %s", columnXML)
			}
			column := 1
			for _, span := range tc.merges {
				for _, row := range []int{3, 4} {
					want := fmt.Sprintf(`<mergeCell ref="%s%d:%s%d"/>`, xlsxColumnName(column), row, xlsxColumnName(column+span-1), row)
					if !strings.Contains(worksheet, want) {
						t.Fatalf("XLSX missing %s", want)
					}
				}
				column += span
			}

			renderer := newPDFRenderer(document.Title)
			renderer.drawFieldRows(sections[0].FieldRows)
			page := string(renderer.finish()[0])
			x := pdfPageMargin
			for _, span := range tc.spans {
				width := pdfContentWidth * span / 24
				want := fmt.Sprintf("%.2f %.2f %.2f %.2f re", x, pdfPageHeight-pdfPageMargin-22, width, 22.0)
				if !strings.Contains(page, want) {
					t.Fatalf("PDF missing exact cell rectangle %q", want)
				}
				x += width
			}
		})
	}
}

func TestFractionalXLSXGridScalesWholeSheetButNotOtherSheets(t *testing.T) {
	fields := fractionalFormFields(t, []float64{4.8, 4.8, 4.8, 4.8, 4.8, 18, 3, 3, 0})
	sections := exportFormSections(fields, map[string]interface{}{
		"field_0": strings.Repeat("字", 50),
	}, nil, true)
	sections = append(sections, exportSection{
		Title: "明细", Headers: []string{"甲", "乙", "丙"},
		Rows: [][]string{{strings.Repeat("字", 70), "乙值", "丙值"}},
	})
	document := exportDocument{Title: "五等分", Metadata: [][]string{{"编号", "test"}}, Sections: sections}
	legacy := exportDocument{Title: "旧表单", Sections: exportFormSections(fractionalFormFields(t, []float64{8, 8, 8}), nil, nil, true)}
	body, err := renderXLSX([]exportDocument{document, legacy})
	if err != nil {
		t.Fatal(err)
	}
	entries := readZipEntries(t, body)
	var sheet struct {
		Columns []struct {
			Max   int     `xml:"max,attr"`
			Width float64 `xml:"width,attr"`
		} `xml:"cols>col"`
		Rows []struct {
			Number int     `xml:"r,attr"`
			Height float64 `xml:"ht,attr"`
		} `xml:"sheetData>row"`
		Merges []struct {
			Ref string `xml:"ref,attr"`
		} `xml:"mergeCells>mergeCell"`
	}
	if err := xml.Unmarshal([]byte(entries["xl/worksheets/sheet1.xml"]), &sheet); err != nil {
		t.Fatal(err)
	}
	if len(sheet.Columns) != 1 || sheet.Columns[0].Max != 120 || sheet.Columns[0].Width != 0.9 {
		t.Fatalf("unexpected fine grid columns: %#v", sheet.Columns)
	}
	merges := make(map[string]bool)
	for _, merge := range sheet.Merges {
		merges[merge.Ref] = true
	}
	for _, want := range []string{
		"A1:DP1", "A2:Y2", "Z2:DP2", "A4:DP4",
		"A7:CL7", "CM7:DA7", "DB7:DP7", "A9:DP9", "A10:DP10",
		"A12:DP12", "A13:AN13", "AO13:CB13", "CC13:DP13",
		"A14:AN14", "AO14:CB14", "CC14:DP14",
	} {
		if !merges[want] {
			t.Errorf("missing whole-sheet merge %s", want)
		}
	}
	// Grid refinement must not make the text-height estimate five times wider.
	for _, row := range sheet.Rows {
		if (row.Number == 6 || row.Number == 14) && row.Height != 56 {
			t.Errorf("long text row %d height = %v, want 56", row.Number, row.Height)
		}
	}
	legacySheet := entries["xl/worksheets/sheet2.xml"]
	for _, want := range []string{`<col min="1" max="24" width="4.5" customWidth="1"/>`, `ref="A1:X1"`, `ref="A3:H3"`, `ref="I3:P3"`, `ref="Q3:X3"`} {
		if !strings.Contains(legacySheet, want) {
			t.Errorf("legacy sheet changed: missing %s", want)
		}
	}
}

func fractionalFormFields(t *testing.T, spans []float64) []workflowcore.FormField {
	t.Helper()
	fields := make([]workflowcore.FormField, len(spans))
	for i, span := range spans {
		payload := fmt.Sprintf(`{"key":"field_%d","label":"F%d","type":"text","span":%g}`, i, i, span)
		if err := json.Unmarshal([]byte(payload), &fields[i]); err != nil {
			t.Fatalf("decode fractional field: %v", err)
		}
	}
	return fields
}

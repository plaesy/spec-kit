package validate

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

// XlsxResult carries the findings of Xlsx, mirroring the [INFO]/[OK] lines
// printed by validate-xlsx.sh's embedded openpyxl check.
type XlsxResult struct {
	Sheets []string
}

type workbookXML struct {
	Sheets struct {
		Sheet []struct {
			Name    string `xml:"name,attr"`
			SheetID string `xml:"sheetId,attr"`
			RID     string `xml:"id,attr"`
		} `xml:"sheet"`
	} `xml:"sheets"`
}

type workbookRelsXML struct {
	Relationship []struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

// Xlsx ports validate-xlsx.sh's structural check:
//  1. the file must be a well-formed OOXML zip containing xl/workbook.xml
//     (bash: "openpyxl can open it")
//  2. xl/workbook.xml must be well-formed XML and list at least one sheet
//     (bash: "workbook has no sheets")
//  3. every name in expectedSheets must appear in the workbook (bash:
//     "expected sheet '%s' not found")
//  4. each sheet's worksheet XML part must be well-formed (bash accesses
//     ws.max_row/max_column, which requires openpyxl to have parsed it)
//
// The bash script's optional LibreOffice render step has no Go-stdlib
// equivalent and is not attempted; see ooxml.go.
func Xlsx(path string, expectedSheets []string) (*XlsxResult, error) {
	zr, err := openOOXML(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	if findPart(zr, "[Content_Types].xml") == nil {
		return nil, fmt.Errorf("not a well-formed OOXML file: missing [Content_Types].xml")
	}

	wbData, err := requirePart(zr, "xl/workbook.xml")
	if err != nil {
		return nil, fmt.Errorf("not a valid xlsx: %w", err)
	}
	var wb workbookXML
	if err := xml.Unmarshal(wbData, &wb); err != nil {
		return nil, fmt.Errorf("xl/workbook.xml is not well-formed XML: %w", err)
	}

	res := &XlsxResult{}
	for _, s := range wb.Sheets.Sheet {
		res.Sheets = append(res.Sheets, s.Name)
	}
	if len(res.Sheets) == 0 {
		return res, fmt.Errorf("workbook has no sheets")
	}

	for _, want := range expectedSheets {
		found := false
		for _, have := range res.Sheets {
			if have == want {
				found = true
				break
			}
		}
		if !found {
			return res, fmt.Errorf("expected sheet '%s' not found in %v", want, res.Sheets)
		}
	}

	// Resolve each sheet's r:id to its worksheet part via
	// xl/_rels/workbook.xml.rels, then confirm that part is well-formed XML
	// (mirrors accessing ws.max_row/max_column, which forces openpyxl to
	// have parsed the sheet).
	relsByID := map[string]string{}
	if relsData, err := requirePart(zr, "xl/_rels/workbook.xml.rels"); err == nil {
		var rels workbookRelsXML
		if err := xml.Unmarshal(relsData, &rels); err == nil {
			for _, r := range rels.Relationship {
				relsByID[r.ID] = r.Target
			}
		}
	}

	for i, s := range wb.Sheets.Sheet {
		target := relsByID[s.RID]
		if target == "" {
			// Fall back to the conventional sheetN.xml naming when rels are
			// absent or unresolvable.
			target = "worksheets/sheet" + strconv.Itoa(i+1) + ".xml"
		}
		partName := "xl/" + normalizeRelTarget(target)
		f := findPart(zr, partName)
		if f == nil {
			continue // best-effort; sheet presence already validated above
		}
		rc, err := f.Open()
		if err != nil {
			return res, fmt.Errorf("cannot open %s: %w", partName, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return res, fmt.Errorf("cannot read %s: %w", partName, err)
		}
		if err := wellFormedXML(data); err != nil {
			return res, fmt.Errorf("worksheet part for sheet %q is not well-formed XML: %w", s.Name, err)
		}
	}

	return res, nil
}

// normalizeRelTarget strips a leading "/xl/" or "./" that some producers use
// in workbook.xml.rels Target attributes, so the result can be joined onto
// the "xl/" prefix uniformly.
func normalizeRelTarget(target string) string {
	for _, prefix := range []string{"/xl/", "xl/", "./"} {
		if len(target) > len(prefix) && target[:len(prefix)] == prefix {
			return target[len(prefix):]
		}
	}
	return target
}

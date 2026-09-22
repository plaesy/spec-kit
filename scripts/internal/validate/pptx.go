package validate

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// PptxResult carries the findings of Pptx, mirroring the [INFO]/[WARN]/[OK]
// lines printed by validate-pptx.sh's embedded python-pptx check.
type PptxResult struct {
	SlideCount    int
	SlidesNoTitle []int // 1-based slide numbers with no title placeholder
}

type presentationXML struct {
	SldIDLst struct {
		SldID []struct{} `xml:"sldIdLst>sldId"`
	}
}

// Pptx ports validate-pptx.sh's structural check:
//  1. the file must be a well-formed OOXML zip containing
//     ppt/presentation.xml (bash: "python-pptx can open it")
//  2. ppt/presentation.xml must be well-formed XML, and its slide count is
//     taken from the <p:sldIdLst> entries (bash: prs.slides / len)
//  3. each slide's XML part is checked for a title placeholder
//     (<p:ph type="title"/> or "ctrTitle"); slides without one are reported
//     the same way the bash script prints "[WARN] slide N has no title
//     placeholder" (non-fatal)
//  4. if expectedSlides > 0, the slide count must match exactly (bash:
//     "expected N slides, got M")
//
// The bash script's optional LibreOffice render step has no Go-stdlib
// equivalent and is not attempted; see ooxml.go.
func Pptx(path string, expectedSlides int) (*PptxResult, error) {
	zr, err := openOOXML(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	if findPart(zr, "[Content_Types].xml") == nil {
		return nil, fmt.Errorf("not a well-formed OOXML file: missing [Content_Types].xml")
	}

	presData, err := requirePart(zr, "ppt/presentation.xml")
	if err != nil {
		return nil, fmt.Errorf("not a valid pptx: %w", err)
	}
	var pres presentationXML
	if err := xml.Unmarshal(presData, &pres); err != nil {
		return nil, fmt.Errorf("ppt/presentation.xml is not well-formed XML: %w", err)
	}

	res := &PptxResult{SlideCount: len(pres.SldIDLst.SldID)}

	// Collect the numbered slide parts (ppt/slides/slideN.xml) present in
	// the archive, in slide-number order, and check each for a title
	// placeholder.
	slideNums := make([]int, 0)
	slideData := make(map[int][]byte)
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "ppt/slides/slide") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		numStr := strings.TrimSuffix(strings.TrimPrefix(f.Name, "ppt/slides/slide"), ".xml")
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return res, fmt.Errorf("cannot open %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return res, fmt.Errorf("cannot read %s: %w", f.Name, err)
		}
		if err := wellFormedXML(data); err != nil {
			return res, fmt.Errorf("%s is not well-formed XML: %w", f.Name, err)
		}
		slideNums = append(slideNums, n)
		slideData[n] = data
	}
	sort.Ints(slideNums)

	for _, n := range slideNums {
		if !hasTitlePlaceholder(slideData[n]) {
			res.SlidesNoTitle = append(res.SlidesNoTitle, n)
		}
	}

	if expectedSlides > 0 && res.SlideCount != expectedSlides {
		return res, fmt.Errorf("expected %d slides, got %d", expectedSlides, res.SlideCount)
	}
	return res, nil
}

type phElement struct {
	Type string `xml:"type,attr"`
}

// hasTitlePlaceholder reports whether the slide XML contains a <p:ph
// type="title"/> or type="ctrTitle" element, matching python-pptx's notion
// of slide.shapes.title.
func hasTitlePlaceholder(data []byte) bool {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return false
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "ph" {
			continue
		}
		var ph phElement
		for _, a := range se.Attr {
			if a.Name.Local == "type" {
				ph.Type = a.Value
			}
		}
		if ph.Type == "title" || ph.Type == "ctrTitle" {
			return true
		}
	}
}

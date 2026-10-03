package validate

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var jinjaTagRE = regexp.MustCompile(`\{\{.*?\}\}`)

// DocxResult carries the findings of Docx, mirroring the [INFO]/[OK] lines
// printed by validate-docx.sh's embedded python-docx check.
type DocxResult struct {
	Paragraphs int
	Tables     int
	Leftover   []string // paragraph texts still containing {{ ... }}
}

// Docx ports validate-docx.sh's structural check:
//  1. the file must be a well-formed OOXML zip containing word/document.xml
//     (bash: "python-docx can open it")
//  2. word/document.xml must be well-formed XML
//  3. the document must have at least one paragraph or table (bash:
//     "document has no paragraphs or tables")
//  4. no paragraph may contain a leftover unfilled {{ }} template tag (bash:
//     "leftover unfilled template tags found")
//
// The bash script's optional LibreOffice render step has no Go-stdlib
// equivalent and is not attempted; see ooxml.go.
func Docx(path string) (*DocxResult, error) {
	zr, err := openOOXML(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	if findPart(zr, "[Content_Types].xml") == nil {
		return nil, fmt.Errorf("not a well-formed OOXML file: missing [Content_Types].xml")
	}
	data, err := requirePart(zr, "word/document.xml")
	if err != nil {
		return nil, fmt.Errorf("not a valid docx: %w", err)
	}

	res, err := parseDocxDocument(data)
	if err != nil {
		return nil, fmt.Errorf("word/document.xml is not well-formed XML: %w", err)
	}

	if res.Paragraphs == 0 && res.Tables == 0 {
		return res, fmt.Errorf("document has no paragraphs or tables")
	}
	if len(res.Leftover) > 0 {
		return res, fmt.Errorf("leftover unfilled template tags found: %v", res.Leftover)
	}
	return res, nil
}

// parseDocxDocument walks word/document.xml with a streaming XML decoder,
// counting <w:p> paragraphs and <w:tbl> tables, and collecting the joined
// text of any paragraph whose runs contain a {{ ... }} tag.
func parseDocxDocument(data []byte) (*DocxResult, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	res := &DocxResult{}

	var textStack []*strings.Builder
	inText := false

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return res, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				res.Paragraphs++
				textStack = append(textStack, &strings.Builder{})
			case "tbl":
				res.Tables++
			case "t":
				inText = true
			}
		case xml.CharData:
			if inText && len(textStack) > 0 {
				textStack[len(textStack)-1].Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				if n := len(textStack); n > 0 {
					text := textStack[n-1].String()
					textStack = textStack[:n-1]
					if jinjaTagRE.MatchString(text) {
						res.Leftover = append(res.Leftover, text)
					}
				}
			}
		}
	}
	return res, nil
}

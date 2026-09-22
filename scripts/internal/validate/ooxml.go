// Package validate ports the OOXML (docx/pptx/xlsx) and Plaesy-memory
// self-containment validators from scripts/bash/validate-*.sh and
// scripts/bash/plaesy-validate-memory.sh.
//
// The bash originals validated docx/pptx/xlsx structure by shelling out to
// python-docx / python-pptx / openpyxl and then, optionally, headless
// rendering the file via LibreOffice (soffice) to catch corruption the
// object model alone would not surface. Neither python nor LibreOffice is
// available to a Go stdlib port, so this package validates the same
// documents directly against the OOXML container format instead: docx/pptx/
// xlsx files are ZIP archives holding XML parts. We open the zip with
// archive/zip, confirm the required parts are present, and parse each with
// encoding/xml to confirm it is well-formed — this is a stricter structural
// check than "python could open it" for the cases it covers, but it cannot
// reproduce the LibreOffice render step, which is skipped with a notice
// (matching the bash behavior when soffice is not installed).
package validate

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

// openOOXML opens path as a zip archive and returns it, erroring with a
// message matching the bash scripts' "File not found" / open-failure checks.
func openOOXML(path string) (*zip.ReadCloser, error) {
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("not a well-formed OOXML (zip) file: %w", err)
	}
	return zr, nil
}

// findPart returns the *zip.File for name, or nil if absent.
func findPart(zr *zip.ReadCloser, name string) *zip.File {
	for _, f := range zr.File {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// requirePart opens and reads the named required OOXML part, erroring if it
// is missing.
func requirePart(zr *zip.ReadCloser, name string) ([]byte, error) {
	f := findPart(zr, name)
	if f == nil {
		return nil, fmt.Errorf("missing required part: %s", name)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("cannot open part %s: %w", name, err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("cannot read part %s: %w", name, err)
	}
	return data, nil
}

// wellFormedXML reports whether data parses as well-formed XML, by walking
// every token to end-of-document. Used where a part only needs a
// syntactic-validity check rather than being unmarshaled into a struct.
func wellFormedXML(data []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// NoRenderCheckNotice mirrors the bash scripts' "LibreOffice (soffice) not
// found — skipping render check" message: this Go port never attempts the
// render step (no LibreOffice dependency), so it always prints the same
// informational notice instead of silently doing less than the bash version.
func NoRenderCheckNotice() string {
	return "[INFO] Headless-render check (LibreOffice) is not available in this Go port — skipping. Structural validation via OOXML zip + XML parsing was performed instead."
}

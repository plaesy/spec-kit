package mdlint

import (
	"fmt"
	"path"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// The three GitHub accessibility rules. They live in the npm-only
// @github/markdownlint-github package, and npm rule plugins are unreachable from
// Go — so the accessibility coverage that config had is reimplemented here rather
// than lost in the migration.
const (
	// GHAEmptyAltText mirrors "no-empty-alt-text".
	GHAEmptyAltText = "GHA001"
	// GHADefaultAltText mirrors "no-default-alt-text" (alt text is just the file name).
	GHADefaultAltText = "GHA002"
	// GHAGenericLinkText mirrors "no-generic-link-text".
	GHAGenericLinkText = "GHA003"
)

// genericLinkTexts is the phrase list from the GitHub ruleset. Comparison is
// case-insensitive and ignores surrounding punctuation.
var genericLinkTexts = []string{
	"click here", "here", "link", "more", "read more", "this", "learn more", "this link", "see more",
}

func checkGHA001(d *doc, _ *ruleSet) []violation {
	var out []violation
	eachImage(d, func(img *ast.Image) {
		if strings.TrimSpace(imageAlt(img, d.source)) == "" {
			out = append(out, violation{GHAEmptyAltText, d.lineOf(img), 1, "image has no alt text"})
		}
	})
	return out
}

func checkGHA002(d *doc, _ *ruleSet) []violation {
	var out []violation
	eachImage(d, func(img *ast.Image) {
		alt := strings.TrimSpace(imageAlt(img, d.source))
		if alt == "" {
			return // already reported by GHA001
		}
		base := path.Base(string(img.Destination))
		stem := strings.TrimSuffix(base, path.Ext(base))
		// Both "logo.png" and "logo" count: neither tells a screen-reader user
		// anything about the image.
		if alt == base || alt == stem {
			out = append(out, violation{GHADefaultAltText, d.lineOf(img), 1,
				fmt.Sprintf("alt text repeats the image file name (%s)", base)})
		}
	})
	return out
}

func checkGHA003(d *doc, _ *ruleSet) []violation {
	var out []violation
	d.walk(func(node ast.Node) bool {
		link, ok := node.(*ast.Link)
		if !ok {
			return true
		}
		text := strings.ToLower(strings.TrimSpace(nodeText(link, d.source)))
		text = strings.Trim(text, ".,:;!?'\"()[]")
		for _, generic := range genericLinkTexts {
			if text == generic {
				out = append(out, violation{GHAGenericLinkText, d.lineOf(link), 1,
					fmt.Sprintf("link text %q does not describe the destination", text)})
				return false
			}
		}
		return true
	})
	return out
}

// eachImage visits every image node in the document.
func eachImage(d *doc, fn func(*ast.Image)) {
	d.walk(func(node ast.Node) bool {
		if img, ok := node.(*ast.Image); ok {
			fn(img)
		}
		return true
	})
}

// imageAlt concatenates an image's alt text.
func imageAlt(img *ast.Image, source []byte) string {
	return nodeText(img, source)
}

// nodeText concatenates the textual children of a node. The source bytes are
// required: goldmark text segments store offsets, not strings.
// nodeText concatenates a node's inline text. goldmark's inline nodes are a
// tree, not a flat list: a heading that is entirely inline markup — `### `a
// target“ — has no direct *ast.Text with a usable segment, so reading only
// direct children returns "" and every such heading looks identical to every
// other. That produced "duplicate heading """ messages naming nothing, and made
// two unrelated headings collide.
func nodeText(node ast.Node, source []byte) string {
	var b strings.Builder
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			switch t := child.(type) {
			case *ast.Text:
				// A *ast.Text with children is a container, not a leaf; taking its
				// own segment as well would double-count the text underneath.
				if child.FirstChild() != nil {
					walk(child)
					continue
				}
				b.Write(t.Segment.Value(source))
			case *ast.String:
				b.Write(t.Value)
			case *ast.AutoLink:
				b.Write(t.URL(source))
			default:
				walk(child)
			}
		}
	}
	walk(node)
	return b.String()
}

package graph

import (
	"regexp"
	"strings"
)

// extractEdges runs every structural-extraction pass (2a-2g in the bash
// script) over the given files and returns EXTRACTED edges. known/ntype let
// each pass check whether a resolved candidate is an actual node.
func extractEdges(repoRootSlash string, files []string, known map[string]bool) []Edge {
	var edges []Edge
	add := func(src, tgt, typ string) {
		if tgt != "" && tgt != src {
			edges = append(edges, Edge{Source: src, Target: tgt, Type: typ, Confidence: "EXTRACTED"})
		}
	}

	// Pre-index Go modules and Dart packages once (2f-go / 2g-dart helpers).
	gomodPath := scanGoModules(repoRootSlash, known)
	goDirFiles := indexGoDirFiles(files)
	dartPkgLib := scanDartPubspecs(repoRootSlash, known)
	javaIndex := indexJavaFiles(files)

	for _, rel := range files {
		content := readAll(joinRoot(repoRootSlash, rel))
		if content == "" {
			continue
		}
		ext := extOf(rel)
		switch ext {
		case ".md":
			for _, m := range mdLinkRE.FindAllStringSubmatch(content, -1) {
				add(rel, resolveRelLink(repoRootSlash, rel, m[1], known), "references")
			}
		case ".ps1":
			for _, m := range ps1DotSourceRE.FindAllStringSubmatch(content, -1) {
				add(rel, resolveRelLink(repoRootSlash, rel, cleanPS1Target(m[1]), known), "sources")
			}
			for _, m := range ps1CallRE.FindAllStringSubmatch(content, -1) {
				add(rel, resolveRelLink(repoRootSlash, rel, cleanPS1Target(m[1]), known), "calls")
			}
		case ".sh":
			for _, m := range shSourceRE.FindAllStringSubmatch(content, -1) {
				add(rel, resolveRelLink(repoRootSlash, rel, m[1], known), "sources")
			}
		case ".js", ".jsx", ".ts", ".tsx":
			for _, m := range jsImportRE.FindAllStringSubmatch(content, -1) {
				spec := firstNonEmpty(m[1], m[2])
				if !strings.HasPrefix(spec, ".") {
					continue
				}
				add(rel, resolveJSImport(repoRootSlash, rel, spec, known), "imports")
			}
		case ".py":
			for _, line := range strings.Split(content, "\n") {
				m := pyImportRE.FindStringSubmatch(line)
				if m == nil {
					continue
				}
				spec := firstNonEmpty(m[1], m[2])
				add(rel, resolvePyImport(repoRootSlash, rel, spec, known), "imports")
			}
		case ".go":
			if len(gomodPath) > 0 {
				for _, line := range strings.Split(content, "\n") {
					m := goImportLineRE.FindStringSubmatch(line)
					if m == nil {
						continue
					}
					for _, tfile := range resolveGoImport(rel, m[1], gomodPath, goDirFiles) {
						add(rel, tfile, "imports")
					}
				}
			}
		case ".dart":
			for _, m := range dartImportRE.FindAllStringSubmatch(content, -1) {
				spec := m[1]
				if strings.HasPrefix(spec, "dart:") {
					continue
				}
				add(rel, resolveDartImport(repoRootSlash, rel, spec, dartPkgLib, known), "imports")
			}
		case ".java":
			for _, line := range strings.Split(content, "\n") {
				m := javaImportRE.FindStringSubmatch(line)
				if m == nil {
					continue
				}
				spec := m[1]
				if strings.HasSuffix(spec, "*") {
					continue
				}
				add(rel, javaIndex[spec], "imports")
			}
		}
	}
	return edges
}

func extOf(rel string) string {
	if i := strings.LastIndexByte(rel, '.'); i >= 0 {
		return strings.ToLower(rel[i:])
	}
	return ""
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// --- 2a) markdown reference links ---
var mdLinkRE = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)\)`)

// --- 2b/2c) PowerShell dot-source / call ---
var (
	ps1DotSourceRE = regexp.MustCompile(`(?m)(?:^|\s)\.\s+["']?([^"'\r\n]+\.ps1)`)
	ps1CallRE      = regexp.MustCompile(`&\s+["']([^"']+\.ps1)`)
)

func cleanPS1Target(t string) string {
	t = strings.TrimPrefix(t, `$PSScriptRoot\`)
	t = strings.TrimPrefix(t, `$PSScriptRoot/`)
	t = strings.ReplaceAll(t, `\`, "/")
	return t
}

// --- 2d) shell source ---
var shSourceRE = regexp.MustCompile(`(?m)(?:^|\s)(?:source|\.)\s+["']?([^"'\r\n]+\.sh)`)

// --- 2e) JS/TS import specs ---
var jsImportRE = regexp.MustCompile(`import\s+.*?from\s+['"]([^'"]+)['"]|require\(['"]([^'"]+)['"]\)`)

func resolveJSImport(repoRootSlash, rel, spec string, known map[string]bool) string {
	fromDir := repoRootSlash
	if i := strings.LastIndexByte(rel, '/'); i >= 0 {
		fromDir = repoRootSlash + "/" + rel[:i]
	}
	candidates := []string{spec, spec + ".js", spec + ".jsx", spec + ".ts", spec + ".tsx", spec + "/index.js", spec + "/index.ts"}
	for _, cand := range candidates {
		resolved := normalizePath(fromDir, cand)
		prefix := repoRootSlash + "/"
		if !strings.HasPrefix(resolved+"/", prefix) {
			continue
		}
		relc := strings.TrimPrefix(resolved, prefix)
		if known[relc] {
			return relc
		}
	}
	return ""
}

// --- 2f) Python import specs ---
var pyImportRE = regexp.MustCompile(`^\s*(?:from\s+([\w.]+)(?:\s+import)|import\s+([\w.]+))`)

func resolvePyImport(repoRootSlash, rel, spec string, known map[string]bool) string {
	if spec == "" || isAllDots(spec) {
		return ""
	}
	fromDir := repoRootSlash
	if i := strings.LastIndexByte(rel, '/'); i >= 0 {
		fromDir = repoRootSlash + "/" + rel[:i]
	}
	specPath := strings.ReplaceAll(spec, ".", "/")
	resolved := normalizePath(fromDir, specPath+".py")
	prefix := repoRootSlash + "/"
	if !strings.HasPrefix(resolved+"/", prefix) {
		return ""
	}
	relc := strings.TrimPrefix(resolved, prefix)
	if known[relc] {
		return relc
	}
	return ""
}

func isAllDots(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c != '.' {
			return false
		}
	}
	return true
}

// --- 2f-go) Go import specs: package (dir) fan-out ---
var goImportLineRE = regexp.MustCompile(`^\s*"([^"]+)"\s*$`)
var goModuleLineRE = regexp.MustCompile(`(?m)^module\s+(\S+)`)

// scanGoModules finds every go.mod under the repo and returns
// {gomod-dir(rel-slash, "" for root) -> declared module path}.
func scanGoModules(repoRootSlash string, known map[string]bool) map[string]string {
	out := map[string]string{}
	for rel := range known {
		if strings.HasSuffix(rel, "/go.mod") || rel == "go.mod" {
			content := readAll(joinRoot(repoRootSlash, rel))
			m := goModuleLineRE.FindStringSubmatch(content)
			if m == nil {
				continue
			}
			dir := strings.TrimSuffix(rel, "/go.mod")
			if dir == rel {
				dir = ""
			}
			out[dir] = m[1]
		}
	}
	return out
}

// indexGoDirFiles groups every non-test .go file by its containing dir
// (rel-slash, "" for repo root).
func indexGoDirFiles(files []string) map[string][]string {
	out := map[string][]string{}
	for _, rel := range files {
		if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") {
			continue
		}
		d := ""
		if i := strings.LastIndexByte(rel, '/'); i >= 0 {
			d = rel[:i]
		}
		out[d] = append(out[d], rel)
	}
	return out
}

// resolveGoImport mirrors the 2f-go loop: find the go.mod whose module path
// is the longest prefix match of spec, compute the target directory, and
// return every non-test .go file in it (excluding rel itself).
func resolveGoImport(rel, spec string, gomodPath map[string]string, goDirFiles map[string][]string) []string {
	relDir := ""
	if i := strings.LastIndexByte(rel, '/'); i >= 0 {
		relDir = rel[:i]
	}
	bestDir, bestLen := "", -1
	for gd := range gomodPath {
		if gd == "" || relDir == gd || strings.HasPrefix(relDir, gd+"/") {
			if len(gd) > bestLen {
				bestLen = len(gd)
				bestDir = gd
			}
		}
	}
	if bestLen < 0 {
		return nil
	}
	modpath := gomodPath[bestDir]
	if spec != modpath && !strings.HasPrefix(spec, modpath+"/") {
		return nil
	}
	subpath := strings.TrimPrefix(spec, modpath)
	subpath = strings.TrimPrefix(subpath, "/")
	targetDir := bestDir
	if subpath != "" {
		if targetDir != "" {
			targetDir += "/" + subpath
		} else {
			targetDir = subpath
		}
	}
	var out []string
	for _, tfile := range goDirFiles[targetDir] {
		if tfile != rel {
			out = append(out, tfile)
		}
	}
	return out
}

// --- 2g-dart) Dart import specs ---
var dartImportRE = regexp.MustCompile(`import\s+['"]([^'"]+)['"]`)
var pubspecNameRE = regexp.MustCompile(`(?m)^name:\s*(\S+)`)

// scanDartPubspecs returns {package-name -> lib-dir(rel-slash)}.
func scanDartPubspecs(repoRootSlash string, known map[string]bool) map[string]string {
	out := map[string]string{}
	for rel := range known {
		if strings.HasSuffix(rel, "/pubspec.yaml") || rel == "pubspec.yaml" {
			content := readAll(joinRoot(repoRootSlash, rel))
			m := pubspecNameRE.FindStringSubmatch(content)
			if m == nil {
				continue
			}
			dir := strings.TrimSuffix(rel, "/pubspec.yaml")
			if dir == rel {
				dir = "."
			}
			out[m[1]] = dir + "/lib"
		}
	}
	return out
}

func resolveDartImport(repoRootSlash, rel, spec string, dartPkgLib map[string]string, known map[string]bool) string {
	var resolvedFull string
	if strings.HasPrefix(spec, "package:") {
		pkgspec := strings.TrimPrefix(spec, "package:")
		pkgname, pkgpath := pkgspec, ""
		if i := strings.IndexByte(pkgspec, '/'); i >= 0 {
			pkgname, pkgpath = pkgspec[:i], pkgspec[i+1:]
		}
		libdir, ok := dartPkgLib[pkgname]
		if !ok {
			return ""
		}
		resolvedFull = normalizePath(repoRootSlash+"/"+libdir, pkgpath)
	} else {
		fromDir := repoRootSlash
		if i := strings.LastIndexByte(rel, '/'); i >= 0 {
			fromDir = repoRootSlash + "/" + rel[:i]
		}
		resolvedFull = normalizePath(fromDir, spec)
	}
	prefix := repoRootSlash + "/"
	if !strings.HasPrefix(resolvedFull+"/", prefix) {
		return ""
	}
	relc := strings.TrimPrefix(resolvedFull, prefix)
	if known[relc] {
		return relc
	}
	return ""
}

// --- 2g-java) Java import specs ---
var javaImportRE = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([\w.]+)\s*;`)

// indexJavaFiles maps a dotted class-ish name (rel path with its top-level
// source-root folder stripped, "/" -> ".") to the file.
func indexJavaFiles(files []string) map[string]string {
	out := map[string]string{}
	for _, rel := range files {
		if !strings.HasSuffix(rel, ".java") {
			continue
		}
		i := strings.IndexByte(rel, '/')
		if i < 0 {
			continue
		}
		dotted := rel[i+1:]
		dotted = strings.TrimSuffix(dotted, ".java")
		dotted = strings.ReplaceAll(dotted, "/", ".")
		out[dotted] = rel
	}
	return out
}

package main

import (
	"strings"
	"testing"
)

func TestRootRelativeWarnings(t *testing.T) {
	files := []buildFile{
		{name: "index.html", data: []byte(`<script src="/assets/app.js"></script><link href="./assets/app.css"><img src="https://cdn.example/x.png">`)},
		{name: "assets/app.js", data: []byte(`import("/assets/chunk.js")`)},
		{name: "assets/chunk.js", data: []byte(`console.log("chunk")`)},
		{name: "assets/app.css", data: []byte(`@font-face{src:url(/fonts/x.woff2)}.hero{background:url(//cdn.example/x.png)}`)},
		{name: "fonts/x.woff2", data: []byte("font")},
	}

	warnings := rootRelativeWarnings(files)
	if len(warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %d: %v", len(warnings), warnings)
	}
	for _, want := range []string{`"/assets/app.js"`, `"/assets/chunk.js"`, `"/fonts/x.woff2"`} {
		found := false
		for _, warning := range warnings {
			if strings.Contains(warning, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing warning for %s in %v", want, warnings)
		}
	}
}

func TestRootRelativeWarningsIgnoresUnknownAndRelativePaths(t *testing.T) {
	files := []buildFile{
		{name: "index.html", data: []byte(`<img src="/api/status"><script src="./assets/app.js"></script><a href="/docs/page">x</a>`)},
		{name: "assets/app.js", data: []byte(`fetch("/v1/health")`)},
	}

	if warnings := rootRelativeWarnings(files); len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
}

func TestRootRelativeWarningsDeduplicates(t *testing.T) {
	files := []buildFile{
		{name: "index.html", data: []byte(`<script src="/assets/app.js"></script><script src="/assets/app.js?v=2"></script>`)},
		{name: "assets/app.js", data: []byte("app")},
	}

	if warnings := rootRelativeWarnings(files); len(warnings) != 1 {
		t.Fatalf("expected one deduplicated warning, got %v", warnings)
	}
}

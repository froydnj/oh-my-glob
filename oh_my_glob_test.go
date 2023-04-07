package oh_my_glob

import (
	"testing"
)

func testCase(t *testing.T, glob string, path string, expected bool) {
	glb := Compile(glob)
	actual := glb.Match(path)
	if expected == true && actual == false {
		t.Fatalf("FAILED: expected `%s` to match `%s`\n", glb.original, path)
	} else if expected == false && actual == true {
		t.Fatalf("FAILED: expected `%s` to NOT match `%s`\n", glb.original, path)
	}
}

func benchmarkCompile(b *testing.B, glob string) {
	for i := 0; i < b.N; i++ {
		Compile(glob)
	}
}

func benchmarkMatch(b *testing.B, glob, path string) {
	compiled := Compile(glob)

	for i := 0; i < b.N; i++ {
		compiled.Match(path)
	}
}

func BenchmarkCompileLiteralPaths(b *testing.B) {
	benchmarkCompile(b, "dev/lib/the_cmd/commands/commands.yaml")
}

func BenchmarkLiteralPaths(b *testing.B) {
	path := "dev/lib/the_cmd/commands/commands.yaml"
	benchmarkMatch(b, path, path)
}

func BenchmarkCompileSubdirFromRoot(b *testing.B) {
	benchmarkCompile(b, "**/*.yaml")
}

func BenchmarkSubdirFromRoot(b *testing.B) {
	benchmarkMatch(b, "**/*.yaml", "dev/lib/the_cmd/commands/commands.yaml")
}

func TestBasicGlob(t *testing.T) {
	// basic string equality
	testCase(t, "what", "what", true)
	testCase(t, "x", "x", true)
	testCase(t, "", "", true)
	testCase(t, "a very long string", "a very long string", true)

	// obviously passing tests
	testCase(t, "wh*", "what", true)
	testCase(t, "*at", "what", true)
	testCase(t, "w*t", "what", true)

	// wildcard, baybee
	testCase(t, "*", "what", true)

	// obviously failing cases
	testCase(t, "wh*", "que", false)
	testCase(t, "wh*", "hut", false)
	testCase(t, "*at", "where", false)
	testCase(t, "w*t", "qat", false)
	testCase(t, "w*t", "where", false)
}

func TestNoWildcards(t *testing.T) {
	// paths with no wildcards
	testCase(t, "foo/bar", "foo/bar", true)
	testCase(t, "foo/bar", "foo/this/bar", false)
	testCase(t, "foo/bar", "foo", false)
}

func TestSingleStar(t *testing.T) {
	// paths with single-star wildcards at start
	testCase(t, "*/bar", "foo/bar", true)
	testCase(t, "*/bar", "baz", false)
	testCase(t, "*/bar", "foo/baz", false)
	testCase(t, "*/bar", "foo/bar/baz", false)

	// paths with single-star wildcards at end
	testCase(t, "foo/*", "foo/bar", true)
	testCase(t, "foo/*", "foo", false)
	testCase(t, "foo/*", "baz/bar", false)
	testCase(t, "foo/*", "foo/baz/bar", false)

	// paths with single-star wildcards in the middle
	testCase(t, "foo/*/bar", "foo/this/bar", true)
	testCase(t, "foo/*/bar", "foo/something-else/bar", true)
	testCase(t, "foo/*/bar", "foo/this/other", false)
	testCase(t, "foo/*/bar", "other/this/bar", false)
	testCase(t, "foo/*/bar", "foo/this/that/bar", false)
}

func TestDoubleStar(t *testing.T) {
	// paths with double-star wildcards at start
	testCase(t, "**/bar", "foo/bar", true)
	testCase(t, "**/bar", "bar", true)
	testCase(t, "**/bar", "foo/baz", false)
	testCase(t, "**/bar", "foo/bar/bar", true)

	// paths with double-star wildcards at end
	testCase(t, "foo/**", "foo/bar", true)
	testCase(t, "foo/**", "foo", true)
	testCase(t, "foo/**", "baz/bar", false)
	testCase(t, "foo/**", "foo/baz/bar", true)

	// paths with double-star wildcards in the middle
	testCase(t, "foo/**/bar", "foo/bar", true)
	testCase(t, "foo/**/bar", "foo/this/bar", true)
	testCase(t, "foo/**/bar", "foo/this/that/bar", true)
	testCase(t, "foo/**/bar", "foo/this/that/the-other/bar", true)
	testCase(t, "foo/**/bar", "this/that/the-other/bar", false)
	testCase(t, "foo/**/bar", "foo/this/that/the-other", false)
	testCase(t, "foo/**/bar", "this/that/the-other", false)
}

func TestCombined(t *testing.T) {
	// combined double-star and single-star
	testCase(t, "config/**/*.conf", "config/foo.conf", true)
	testCase(t, "config/**/*.conf", "config/this/foo.conf", true)
	testCase(t, "config/**/*.conf", "config/this/that/foo.conf", true)
	testCase(t, "config/**/*.conf", "something/else.conf", false)
}

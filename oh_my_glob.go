package oh_my_glob

import (
	"log"
	"strings"
)

// "enum" for parts, below.
const (
	// Something that is not a literal `*` or `**`, but makes no other
	// promises about its contents.
	literal    = iota
	singleStar = iota
	doubleStar = iota
)

type part struct {
	kind int8
	// only set if kind == literal
	no_stars bool
	// only set if kind == literal
	lit string
}

var starstar part = part{
	kind: doubleStar,
	lit:  "",
}

var star part = part{
	kind: singleStar,
	lit:  "",
}

type Glob struct {
	// this is for pretty-printing the glob
	original string
	// we pre-break the glob at `/` boundaries for less processing
	// later
	parts []part
}

func Compile(glob string) Glob {
	if glob == "" {
		return Glob{
			original: "",
			parts:    nil,
		}
	}

	fragments := strings.Split(glob, "/")
	parts := make([]part, 0, len(fragments))
	for _, fragment := range fragments {
		if fragment == "**" {
			parts = append(parts, starstar)
		} else if fragment == "*" {
			parts = append(parts, star)
		} else {
			parts = append(parts, part{
				kind:     literal,
				lit:      fragment,
				no_stars: strings.IndexByte(fragment, '*') == -1,
			})
		}
	}
	return Glob{
		original: glob,
		parts:    parts,
	}
}

// this code is borrowed directly from Russ Cox's research page,
// albeit with the single-character wildcard removed (since I don't
// believe we need it.) For more details:
//
//	https://research.swtch.com/glob
func match(pattern, name string) bool {
	px := 0
	nx := 0
	nextPx := 0
	nextNx := 0
	for px < len(pattern) || nx < len(name) {
		if px < len(pattern) {
			c := pattern[px]
			switch c {
			default:
				if nx < len(name) && name[nx] == c {
					px++
					nx++
					continue
				}

			case '*':
				// zero-or-more-character wildcard
				// Try to match at nx.
				// If that doesn't work out,
				// restart at nx+1 next.
				nextPx = px
				nextNx = nx + 1
				px++
				continue
			}
		}
		// Mismatch. Maybe restart.
		if 0 < nextNx && nextNx <= len(name) {
			px = nextPx
			nx = nextNx
			continue
		}
		return false
	}
	// Matched all of pattern to all of name. Success.
	return true
}

func (g *Glob) Match(path string) bool {
	// `px` is the index into the current path part
	px := 0
	// `nx` is the index into the string, which will change in
	// strides based on where we find `/` characters
	nx := 0
	// these are used for backtracking in the case of `**`,
	// c.f. the Russ Cox page linked above
	nextPx := 0
	nextNx := 0

	for px < len(g.parts) || nx < len(path) {
		// This is a little hairy: `incrNx` is going to be the
		// "next" `nx` value, and we'll only ever need it if
		// `nx < len(path)` above. `nx` should always point to
		// the beginning of a path segment _or_ to the end of
		// the path. We start searching from the current `nx`
		// and find the next '/' character
		incrNx := 0
		if nx < len(path) {
			tx := strings.IndexByte(path[nx:], '/')

			if tx < 0 {
				incrNx = len(path)
			} else {
				incrNx = nx + tx + 1
			}
		}

		if px < len(g.parts) {
			c := g.parts[px]
			switch c.kind {
			case literal:
				if nx < len(path) {
					// find the next substring
					var chunk string
					if incrNx == len(path) {
						chunk = path[nx:]
					} else {
						chunk = path[nx : incrNx-1]
					}

					if c.no_stars && c.lit == chunk {
						px++
						nx = incrNx
						continue
					} else if match(c.lit, chunk) {
						px++
						nx = incrNx
						continue
					}
				}
			case doubleStar:
				nextPx = px
				nextNx = incrNx
				px++
				continue
			case singleStar:
				if nx < len(path) {
					px++
					nx = incrNx
					continue
				}
			default:
				// this should never happen and
				// indicates a bug in library code
				log.Fatalf("Unexpected compiled glob value")
			}
		}

		if 0 < nextNx && nextNx <= len(path) {
			px = nextPx
			nx = nextNx
			continue
		}
		return false
	}
	return true
}

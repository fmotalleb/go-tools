package glob

import (
	"errors"
	"strings"
	"unsafe"
)

// Matcher represents a compiled glob pattern.
type Matcher struct {
	pattern  string
	segments []segment
}

type segmentType uint8

const (
	segmentLiteral segmentType = iota
	segmentStar
	segmentQuestion
	segmentBraceExpansion
)

type segment struct {
	typ        segmentType
	value      string
	expansions []string // for brace expansions
}

var (
	ErrInvalidPattern = errors.New("invalid glob pattern")
	ErrUnmatchedBrace = errors.New("unmatched brace in pattern")
)

// Compile compiles a glob pattern into a Matcher.
func Compile(pattern string) (*Matcher, error) {
	m := &Matcher{pattern: pattern}
	if err := m.compile(); err != nil {
		return nil, err
	}
	return m, nil
}

// MustCompile compiles a glob pattern and panics on error.
func MustCompile(pattern string) *Matcher {
	m, err := Compile(pattern)
	if err != nil {
		panic(err)
	}
	return m
}

func (m *Matcher) compile() error {
	var segments []segment
	var current strings.Builder

	i := 0
	for i < len(m.pattern) {
		switch m.pattern[i] {
		case '*':
			if current.Len() > 0 {
				segments = append(segments, segment{
					typ:   segmentLiteral,
					value: current.String(),
				})
				current.Reset()
			}
			segments = append(segments, segment{typ: segmentStar})
			i++

		case '?':
			if current.Len() > 0 {
				segments = append(segments, segment{
					typ:   segmentLiteral,
					value: current.String(),
				})
				current.Reset()
			}
			segments = append(segments, segment{typ: segmentQuestion})
			i++

		case '{':
			if current.Len() > 0 {
				segments = append(segments, segment{
					typ:   segmentLiteral,
					value: current.String(),
				})
				current.Reset()
			}

			// Find matching closing brace
			braceEnd := i + 1
			braceDepth := 1
			for braceEnd < len(m.pattern) && braceDepth > 0 {
				switch m.pattern[braceEnd] {
				case '{':
					braceDepth++
				case '}':
					braceDepth--
				}
				braceEnd++
			}

			if braceDepth != 0 {
				return ErrUnmatchedBrace
			}

			// Parse brace expansion
			expansions := globParseBraceExpansion(m.pattern[i+1 : braceEnd-1])
			segments = append(segments, segment{
				typ:        segmentBraceExpansion,
				expansions: expansions,
			})
			i = braceEnd

		default:
			current.WriteByte(m.pattern[i])
			i++
		}
	}

	if current.Len() > 0 {
		segments = append(segments, segment{
			typ:   segmentLiteral,
			value: current.String(),
		})
	}

	m.segments = segments
	return nil
}

func globParseBraceExpansion(content string) []string {
	var result []string
	var current strings.Builder

	braceDepth := 0
	for i := 0; i < len(content); i++ {
		switch content[i] {
		case '{':
			braceDepth++
			current.WriteByte(content[i])
		case '}':
			braceDepth--
			current.WriteByte(content[i])
		case ',':
			if braceDepth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteByte(content[i])
			}
		default:
			current.WriteByte(content[i])
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// Match checks if the input matches the pattern with zero allocations.
func (m *Matcher) Match(input string) bool {
	return m.match(input, 0, 0)
}

func (m *Matcher) match(input string, inputPos, segmentPos int) bool {
	// Base case: both exhausted
	if segmentPos >= len(m.segments) {
		return inputPos >= len(input)
	}

	segment := &m.segments[segmentPos]

	switch segment.typ {
	case segmentLiteral:
		if !globMatchLiteral(input, inputPos, segment.value) {
			return false
		}
		return m.match(input, inputPos+len(segment.value), segmentPos+1)

	case segmentQuestion:
		if inputPos >= len(input) {
			return false
		}
		return m.match(input, inputPos+1, segmentPos+1)

	case segmentStar:
		// Try matching zero or more characters
		for i := inputPos; i <= len(input); i++ {
			if m.match(input, i, segmentPos+1) {
				return true
			}
		}
		return false

	case segmentBraceExpansion:
		// Try each expansion
		for _, expansion := range segment.expansions {
			if globMatchLiteral(input, inputPos, expansion) {
				if m.match(input, inputPos+len(expansion), segmentPos+1) {
					return true
				}
			}
		}
		return false

	default:
		return false
	}
}

// globMatchLiteral performs zero-allocation string prefix matching.
func globMatchLiteral(input string, pos int, literal string) bool {
	if pos+len(literal) > len(input) {
		return false
	}

	// Direct byte comparison without allocation
	inputBytes := stringToBytes(input[pos : pos+len(literal)])
	literalBytes := stringToBytes(literal)

	for i := range literalBytes {
		if inputBytes[i] != literalBytes[i] {
			return false
		}
	}

	return true
}

// stringToBytes converts string to []byte without allocation.
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// MatchString is an alias for Match.
func (m *Matcher) MatchString(s string) bool {
	return m.Match(s)
}

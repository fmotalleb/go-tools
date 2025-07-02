package glob_test

import (
	"fmt"
	"testing"

	"github.com/FMotalleb/go-tools/internal/models"
	"github.com/FMotalleb/go-tools/matcher/glob"
	"github.com/alecthomas/assert/v2"
)

func TestGlobMatcher(t *testing.T) {
	tests := []models.MatchTestCase{
		// Basic literal matching
		{Pattern: "example.com", Input: "example.com", Matches: true},
		{Pattern: "example.com", Input: "test.com", Matches: false},

		// Wildcard (*)
		{Pattern: "*.example.com", Input: "test.example.com", Matches: true},
		{Pattern: "*.example.com", Input: "sub.test.example.com", Matches: true},
		{Pattern: "*.example.com", Input: "example.com", Matches: false},
		{Pattern: "*example*", Input: "testexampletest", Matches: true},
		{Pattern: "*example*", Input: "testexample", Matches: true},
		{Pattern: "*example*", Input: "exampletest", Matches: true},

		// Single character (?)
		{Pattern: "test?.example.com", Input: "test1.example.com", Matches: true},
		{Pattern: "test?.example.com", Input: "test12.example.com", Matches: false},
		{Pattern: "test?.example.com", Input: "test.example.com", Matches: false},

		// Alternation ({})
		{Pattern: "example.com{api,pwa}", Input: "example.comapi", Matches: true},
		{Pattern: "example.com{api,pwa}", Input: "example.compwa", Matches: true},
		{Pattern: "example.com{api,pwa}", Input: "example.comother", Matches: false},
		{Pattern: "example.com{api,pwa}", Input: "example.com", Matches: false},

		// Combination of *, ?, and {}
		{Pattern: "*.example.com{api,pwa}", Input: "test.example.comapi", Matches: true},
		{Pattern: "*.example.com{api,pwa}", Input: "test.example.compwa", Matches: true},
		{Pattern: "*.example.com{api,pwa}", Input: "test.example.comother", Matches: false},
		{Pattern: "*.example.com{api,pwa}", Input: "example.comapi", Matches: false}, // Missing subdomain

		// Edge cases
		{Pattern: "*", Input: "anystring", Matches: true},
		{Pattern: "*", Input: "", Matches: true},
		{Pattern: "", Input: "", Matches: true},   // Empty pattern matches empty string
		{Pattern: "?", Input: "", Matches: false}, // Single character required
		{Pattern: "{a,b,c}", Input: "a", Matches: true},
		{Pattern: "{a,b,c}", Input: "b", Matches: true},
		{Pattern: "{a,b,c}", Input: "d", Matches: false},
	}

	for _, test := range tests {
		t.Run(test.Pattern+"_"+test.Input, func(t *testing.T) {
			matcher, err := glob.Compile(test.Pattern)
			assert.NoError(t, err, "Error compiling pattern")
			assert.Equal(t, test.Matches, matcher.Match(test.Input))
		})
	}
}

func TestGlobMatcherCompileError(t *testing.T) {
	// Test invalid patterns
	invalidPatterns := []string{
		"{unclosed",
		"start{middle",
	}

	for _, pattern := range invalidPatterns {
		t.Run(pattern, func(t *testing.T) {
			_, err := glob.Compile(pattern)
			assert.Error(t, err, "Expected error for invalid pattern")
		})
	}
}

// Benchmark compilation
func BenchmarkCompile(b *testing.B) {
	patterns := []string{
		"example.com",                            // Simple literal
		"*.example.com",                          // Star pattern
		"api-??.example.com",                     // Question marks
		"example.{com,net,org}",                  // Brace expansion
		"*.api-??.{example.com,test.net,dev.io}", // Complex pattern
	}

	for _, pattern := range patterns {
		b.Run(pattern, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := glob.Compile(pattern)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark matching with pre-compiled patterns
func BenchmarkMatch(b *testing.B) {
	benchmarks := []struct {
		name    string
		pattern string
		inputs  []string
	}{
		{
			name:    "Literal",
			pattern: "api.example.com",
			inputs: []string{
				"api.example.com", // match
				"www.example.com", // no match
			},
		},
		{
			name:    "StarPrefix",
			pattern: "*.example.com",
			inputs: []string{
				"api.example.com",     // match
				"www.api.example.com", // match
				"example.com",         // no match
				"api.example.net",     // no match
			},
		},
		{
			name:    "QuestionMark",
			pattern: "api-??.example.com",
			inputs: []string{
				"api-01.example.com",  // match
				"api-99.example.com",  // match
				"api-1.example.com",   // no match
				"api-100.example.com", // no match
			},
		},
		{
			name:    "BraceExpansion",
			pattern: "api.example.{com,net,org}",
			inputs: []string{
				"api.example.com", // match
				"api.example.net", // match
				"api.example.org", // match
				"api.example.io",  // no match
			},
		},
		{
			name:    "Complex",
			pattern: "*.api-??.{example.com,test.net}",
			inputs: []string{
				"subdomain.api-01.example.com", // match
				"subdomain.api-99.test.net",    // match
				"subdomain.api-1.example.com",  // no match
				"api-01.example.com",           // no match
			},
		},
	}

	for _, bm := range benchmarks {
		matcher := glob.MustCompile(bm.pattern)

		for _, input := range bm.inputs {
			name := fmt.Sprintf("%s/%s", bm.name, input)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(input)))

				for i := 0; i < b.N; i++ {
					_ = matcher.Match(input)
				}
			})
		}
	}
}

// Benchmark to verify zero allocations
func BenchmarkMatchZeroAlloc(b *testing.B) {
	matcher := glob.MustCompile("*.api-??.{example.com,test.net}")
	input := "subdomain.api-01.example.com"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !matcher.Match(input) {
			b.Fatal("Expected match")
		}
	}

	// This should report 0 allocs/op
}

// Benchmark different input lengths
func BenchmarkMatchInputLength(b *testing.B) {
	pattern := "*.example.com"
	matcher := glob.MustCompile(pattern)

	inputs := []string{
		"a.example.com",                             // 13 bytes
		"api.example.com",                           // 15 bytes
		"subdomain.example.com",                     // 21 bytes
		"very-long-subdomain.example.com",           // 31 bytes
		"extremely-long-subdomain-name.example.com", // 41 bytes
	}

	for _, input := range inputs {
		b.Run(fmt.Sprintf("len_%d", len(input)), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(input)))

			for i := 0; i < b.N; i++ {
				_ = matcher.Match(input)
			}
		})
	}
}

// Benchmark worst-case scenarios
func BenchmarkMatchWorstCase(b *testing.B) {
	benchmarks := []struct {
		name    string
		pattern string
		input   string
	}{
		{
			name:    "MultipleStars",
			pattern: "*.*.*.*",
			input:   "a.b.c.d.e.f.g.h",
		},
		{
			name:    "DeepBraceExpansion",
			pattern: "api.{v1,v2,v3,v4,v5}.{users,posts,comments}.example.com",
			input:   "api.v3.posts.example.com",
		},
		{
			name:    "LongNonMatch",
			pattern: "*.verylongdomainname.example.com",
			input:   "subdomain.verylongdomainname.example.net", // no match at the end
		},
	}

	for _, bm := range benchmarks {
		matcher := glob.MustCompile(bm.pattern)

		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(bm.input)))

			for i := 0; i < b.N; i++ {
				_ = matcher.Match(bm.input)
			}
		})
	}
}

// Parallel benchmark
func BenchmarkMatchParallel(b *testing.B) {
	matcher := glob.MustCompile("*.api-??.example.com")
	inputs := []string{
		"subdomain.api-01.example.com",
		"subdomain.api-99.example.com",
		"other.api-42.example.com",
		"test.api-00.example.com",
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			input := inputs[i%len(inputs)]
			_ = matcher.Match(input)
			i++
		}
	})
}

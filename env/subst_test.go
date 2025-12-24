package env_test

import (
	"os"
	"testing"

	"github.com/fmotalleb/go-tools/env"
)

func TestSubst(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		input     string
		want      string
		wantPanic bool
	}{
		{
			name:  "basic variable",
			env:   map[string]string{"FOO": "bar"},
			input: "$FOO",
			want:  "bar",
		},
		{
			name:  "basic variable braces",
			env:   map[string]string{"FOO": "bar"},
			input: "${FOO}",
			want:  "bar",
		},
		{
			name:  "missing variable",
			env:   map[string]string{},
			input: "$FOO",
			want:  "$FOO",
		},
		{
			name:  "default value when empty",
			env:   map[string]string{},
			input: "${FOO:-baz}",
			want:  "baz",
		},
		{
			name:  "default value when set",
			env:   map[string]string{"FOO": "bar"},
			input: "${FOO:-baz}",
			want:  "bar",
		},
		{
			name:  "alternate value when set",
			env:   map[string]string{"FOO": "bar"},
			input: "${FOO:+baz}",
			want:  "baz",
		},
		{
			name:  "alternate value when unset",
			env:   map[string]string{},
			input: "${FOO:+baz}",
			want:  "",
		},
		{
			name:      "error when unset",
			env:       map[string]string{},
			input:     "${FOO:?custom error message}",
			wantPanic: true,
		},
		{
			name:  "error when unset",
			env:   map[string]string{"FOO": "bar"},
			input: "${FOO:?custom error message}",
			want:  "bar",
		},
		{
			name:  "escaped dollar sign",
			env:   map[string]string{"FOO": "bar"},
			input: `\$FOO`,
			want:  "$FOO", // no substitution
		},
		{
			name:  "double dollar literal",
			env:   map[string]string{"FOO": "bar"},
			input: `$$FOO`,
			want:  "$FOO", // no substitution
		},
		{
			name:  "mixed text",
			env:   map[string]string{"FOO": "bar"},
			input: "Value=$FOO End",
			want:  "Value=bar End",
		},
		{
			name:  "unterminated brace",
			env:   map[string]string{"FOO": "bar"},
			input: "${FOO",
			want:  "${FOO", // literal return
		},
		{
			name:  "literal: ignore escape",
			env:   map[string]string{},
			input: "\\test text",
			want:  "\\test text",
		},
		{
			name:  "literal: ignore ending",
			env:   map[string]string{},
			input: "finish with $",
			want:  "finish with $",
		},
		{
			name:  "literal: ignore ending",
			env:   map[string]string{},
			input: "finish with ${",
			want:  "finish with ${",
		},
		{
			name:  "literal: not-set braced variable",
			env:   map[string]string{},
			input: "finish with ${INVALID}",
			want:  "finish with ",
		},
		{
			name:  "literal: invalid env with brace ending",
			env:   map[string]string{},
			input: "finish with $INVALID}",
			want:  "finish with $INVALID}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clear env first
			os.Clearenv()
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Fatalf("unexpected panic: %v", r)
					}
				} else if tt.wantPanic {
					t.Fatalf("expected panic but got none")
				}
			}()

			got := env.Subst(tt.input)
			if got != tt.want {
				t.Errorf("Subst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

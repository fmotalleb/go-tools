package env

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"sync"
)

type substOperator struct {
	regex   *regexp.Regexp
	handler func(matches []string) string
}

var (
	substWhenEmpty = substOperator{
		regex: regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?\}`),
		handler: func(matches []string) string {
			varName := matches[1]
			defaultValue := ""
			if len(matches) > 3 {
				defaultValue = matches[3]
			}

			if value := os.Getenv(varName); value != "" {
				return value
			}
			return defaultValue
		},
	}
	substWhenExists = substOperator{
		regex: regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*):(\+([^}]*))?\}`),
		handler: func(matches []string) string {
			varName := matches[1]
			alternateValue := ""
			if len(matches) > 3 {
				alternateValue = matches[3]
			}

			if value := os.Getenv(varName); value != "" {
				return alternateValue
			}
			return ""
		},
	}
	substWhenEmptyError = substOperator{
		regex: regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*):(\?([^}]*))?\}`),
		handler: func(matches []string) string {
			varName := matches[1]
			errorMsg := varName + ": parameter null or not set"
			if len(matches) > 3 && matches[3] != "" {
				errorMsg = matches[3]
			}

			if value := os.Getenv(varName); value != "" {
				return value
			}
			panic(errors.New(errorMsg))
		},
	}
	substBasicEnv = substOperator{
		regex: regexp.MustCompile(`\$(([A-Za-z_][A-Za-z0-9_]*)|\{([A-Za-z_][A-Za-z0-9_]*)\})`),
		handler: func(matches []string) string {
			return os.Getenv(matches[2])
		},
	}
	substPatterns = []substOperator{
		substWhenEmpty,
		substWhenExists,
		substWhenEmptyError,
		substBasicEnv,
	}
)

var escapedEnvSelector = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`(\\\$|\$\$)`)
})

// Subst performs advanced environment variable substitution
func Subst(input string) string {
	// Handle escaped dollar signs by temporarily replacing them, handle ($$ or \$)
	escaped := escapedEnvSelector().ReplaceAllString(input, "§ESCAPED_DOLLAR§")

	// Apply substitutions in order of precedence
	result := escaped
	for _, pattern := range substPatterns {
		result = pattern.regex.ReplaceAllStringFunc(result, func(match string) string {
			matches := pattern.regex.FindStringSubmatch(match)
			if matches == nil {
				return match
			}
			return pattern.handler(matches)
		})
	}

	// Restore escaped dollar signs
	result = strings.ReplaceAll(result, "§ESCAPED_DOLLAR§", "$")

	return result
}

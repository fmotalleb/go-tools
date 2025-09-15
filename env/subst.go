package env

import (
	"bytes"
	"errors"
	"os"
	"regexp"
	"strings"
)

type substOperator struct {
	regex   *regexp.Regexp
	handler func(matches []string) string
}

var (
	substWhenEmpty = substOperator{
		regex: regexp.MustCompile(`^\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?\}$`),
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
		regex: regexp.MustCompile(`^\{([A-Za-z_][A-Za-z0-9_]*):(\+([^}]*))?\}$`),
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
		regex: regexp.MustCompile(`^\{([A-Za-z_][A-Za-z0-9_]*):(\?([^}]*))?\}$`),
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
		regex: regexp.MustCompile(`^(([A-Za-z_][A-Za-z0-9_]*)|\{([A-Za-z_][A-Za-z0-9_]*)\})$`),
		handler: func(matches []string) string {
			name := matches[2]
			if name == "" {
				name = matches[3]
			}
			return os.Getenv(name)
		},
	}
	substPatterns = []substOperator{
		substBasicEnv,
		substWhenEmpty,
		substWhenExists,
		substWhenEmptyError,
	}
)

func Subst(input string) string {
	reader := bytes.NewReader([]byte(input))
	b := new(strings.Builder)

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break // EOF
		}

		switch r {
		case '\\': // Handle escaped $
			next, _, err := reader.ReadRune()
			if err == nil {
				if next == '$' {
					b.WriteRune('$')
				} else {
					b.WriteRune('\\')
					b.WriteRune(next)
				}
			}
		case '$':
			// Peek next rune
			next, _, err := reader.ReadRune()
			if err != nil {
				b.WriteRune('$')
				break
			}
			if next == '$' {
				b.WriteRune('$')
				continue
			}

			// unread so getVar can consume
			_ = reader.UnreadRune()
			b.WriteString(getVar(reader))
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}
func getVar(reader *bytes.Reader) string {
	varName := new(strings.Builder)
	r, _, err := reader.ReadRune()
	if err != nil {
		return ""
	}

	startedWithBrace := false
	if r == '{' {
		startedWithBrace = true
		r, _, err = reader.ReadRune()
		if err != nil || !isVarStart(r) {
			return "" // invalid start
		}
	}
	varName.WriteRune(r)

	for {
		peek, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		if startedWithBrace {
			varName.WriteRune(peek)
			if peek == '}' {
				break
			}
		} else if isVarChar(peek) {
			varName.WriteRune(peek)
		} else {
			_ = reader.UnreadRune()
			break
		}
	}

	vName := varName.String()
	if startedWithBrace && !strings.HasSuffix(vName, "}") {
		return "${" + vName // return literal if not properly closed
	}

	for _, pattern := range substPatterns {
		matches := pattern.regex.FindStringSubmatch(vName)
		if matches != nil {
			return pattern.handler(matches)
		}
	}
	return ""
}

func isVarStart(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		r == '_'
}

func isVarChar(r rune) bool {
	return isVarStart(r) || (r >= '0' && r <= '9')
}

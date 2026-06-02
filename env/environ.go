package env

import (
	"cmp"
	"os"
	"strconv"
	"strings"
	"time"
)

// Or returns environment variable value or first non-empty default.
func Or(key string, def ...string) string {
	if key == "" {
		return cmp.Or(def...)
	}
	items := []string{os.Getenv(key)}
	items = append(items, def...)
	return cmp.Or(items...)
}

// BoolOr returns environment variable as bool or default.
func BoolOr(key string, def ...bool) bool {
	if key == "" {
		return cmp.Or(def...)
	}
	if env := os.Getenv(key); env != "" {
		if val, err := strconv.ParseBool(env); err == nil {
			return val
		}
	}
	return cmp.Or(def...)
}

// IntOr returns environment variable as int or default.
func IntOr(key string, def ...int) int {
	if key == "" {
		return cmp.Or(def...)
	}
	if env := os.Getenv(key); env != "" {
		if val, err := strconv.Atoi(env); err == nil {
			return val
		}
	}
	return cmp.Or(def...)
}

// SliceOr returns environment variable as slice (comma-separated) or default.
func SliceOr(key string, def []string) []string {
	if key == "" {
		return def
	}
	return SliceSeparatorOr(key, ",", def)
}

// SliceOr returns environment variable as slice (comma-separated) or default.
func SliceSeparatorOr(key string, sep string, def []string) []string {
	if env := os.Getenv(key); env != "" {
		return strings.Split(env, sep)
	}
	return def
}

// DurationOr returns environment variable as Duration or default.
func DurationOr(key string, def ...time.Duration) time.Duration {
	if key == "" {
		return cmp.Or(def...)
	}
	if env := os.Getenv(key); env != "" {
		if val, err := time.ParseDuration(env); err == nil {
			return val
		}
	}
	return cmp.Or(def...)
}

func Float64Or(key string, def ...float64) float64 {
	if key == "" {
		return cmp.Or(def...)
	}
	if strVal := os.Getenv(key); strVal != "" {
		if val, err := strconv.ParseFloat(strVal, 64); err == nil {
			return val
		}
	}
	return cmp.Or(def...)
}

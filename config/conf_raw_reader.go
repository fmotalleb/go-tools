package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/FMotalleb/go-tools/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func ReadAndMergeConfig(ctx context.Context, confPath string, includeField ...string) (map[string]any, error) {
	logger := log.FromContext(ctx).
		Named("config-reader")
	currentCtx := log.WithLogger(ctx, logger)
	includeFieldName := "include"
	if len(includeField) != 0 {
		includeFieldName = includeField[0]
	}
	return mergeFromPattern(currentCtx, includeFieldName, confPath, map[string]bool{})
}

func deepMerge(dst, src map[string]any) (map[string]any, error) {
	err := mergo.Merge(&dst, src, mergo.WithAppendSlice)
	return dst, err
}

func parseConfig(ctx context.Context, ext string, reader io.Reader, path string) (map[string]any, error) {
	log := log.
		FromContext(ctx).
		With(zap.String("path", path))
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	v := viper.New()
	v.SetConfigType(ext)
	if err := v.ReadConfig(reader); err != nil {
		log.Error("failed to read config", zap.Error(err))
		return nil, err
	}

	raw := make(map[string]any)
	if err := v.Unmarshal(&raw); err != nil {
		log.Error("failed to unmarshal config", zap.Error(err))
		return nil, err
	}
	return raw, nil
}

func mergeFromPattern(ctx context.Context, includeField string, pattern string, visited map[string]bool) (map[string]any, error) {
	log := log.
		FromContext(ctx).
		With(zap.String("pattern", pattern))
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	var files []string
	if url, err := url.Parse(pattern); err != nil {
		switch url.Scheme {
		case "http", "https":
			log.Debug("pattern parsed as a url with https|http schema")
			files = []string{pattern}
		default:
			log.Debug("pattern parsed as a url but no https|http schema found")
			files = make([]string, 0)
		}
	}
	if len(files) == 0 {
		var err error
		files, err = filepath.Glob(pattern)
		if err != nil {
			log.Error("invalid glob pattern", zap.Error(err))
			return nil, fmt.Errorf("invalid glob pattern: %w", err)
		}
		if len(files) == 0 {
			log.Debug("no config files matched pattern, using it as full path address")
		}
	}

	result := make(map[string]any)
	for _, file := range files {
		innerLog := log.
			With(zap.String("file", file))
		innerLog.Debug("merging config file")
		conf, err := readAndResolveIncludes(ctx, includeField, file, visited)
		if err != nil {
			innerLog.Error("failed to read and merge includes", zap.Error(err))
			continue
		}
		result, err = deepMerge(result, conf)
		if err != nil {
			innerLog.Error("deep merge failed", zap.Error(err))
			continue
		}
		innerLog.Debug("merged successfully")
	}
	return result, nil
}

func readAndResolveIncludes(ctx context.Context, includeField string, path string, visited map[string]bool) (map[string]any, error) {
	log := log.FromContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	absPath := path
	log = log.With(zap.String("path", absPath))
	if u, err := url.Parse(path); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		absPath = u.String()
	} else if p, err := filepath.Abs(path); err == nil {
		absPath = p
	}

	if visited[absPath] {
		log.Warn("circular include detected")
		return make(map[string]any), nil
	}
	log.Info("reading config")
	visited[absPath] = true

	reader, ext, err := readFrom(ctx, path)
	if err != nil {
		return nil, err
	}

	raw, err := parseConfig(ctx, ext, reader, path)
	if err != nil {
		return nil, err
	}
	log.Debug("parsed config")

	raw, err = readIncludedFiles(ctx, includeField, raw, path, visited)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func readIncludedFiles(ctx context.Context, includeField string, raw map[string]any, path string, visited map[string]bool) (map[string]any, error) {
	log := log.
		FromContext(ctx).
		With(zap.String("from", path))
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	if includes, ok := raw[includeField].([]any); ok {
		for _, inc := range includes {
			if pattern, ok := inc.(string); ok {
				log = log.With(zap.String("pattern", pattern))
				log.Info("processing include")
				included, err := mergeFromPattern(ctx, includeField, pattern, visited)
				if err != nil {
					log.Error("failed to process include", zap.Error(err))
					return nil, err
				}
				raw, err = deepMerge(included, raw)
				if err != nil {
					log.Error("peep merge failed during include", zap.Error(err))
					return nil, err
				}
				log.Debug("include merged")
			}
		}
	}
	return raw, nil
}

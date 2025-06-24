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
	log := log.FromContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	v := viper.New()
	v.SetConfigType(ext)
	if err := v.ReadConfig(reader); err != nil {
		log.Error("failed to read config", zap.String("path", path), zap.Error(err))
		return nil, err
	}

	raw := make(map[string]any)
	if err := v.Unmarshal(&raw); err != nil {
		log.Error("failed to unmarshal config", zap.String("path", path), zap.Error(err))
		return nil, err
	}
	return raw, nil
}

func mergeFromPattern(ctx context.Context, includeField string, pattern string, visited map[string]bool) (map[string]any, error) {
	log := log.FromContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Error("invalid glob pattern", zap.String("pattern", pattern), zap.Error(err))
		return nil, fmt.Errorf("invalid glob pattern: %w", err)
	}
	if len(files) == 0 {
		log.Debug("no config files matched pattern, using it as full path address", zap.String("pattern", pattern))
		files = []string{pattern}
	}

	result := make(map[string]any)
	for _, file := range files {
		log.Debug("merging config file", zap.String("file", file))
		conf, err := readAndResolveIncludes(ctx, includeField, file, visited)
		if err != nil {
			log.Error("failed to read and merge includes", zap.String("file", file), zap.Error(err))
			return nil, err
		}
		result, err = deepMerge(result, conf)
		if err != nil {
			log.Error("deep merge failed", zap.String("file", file), zap.Error(err))
			return nil, err
		}
		log.Debug("merged successfully", zap.String("file", file))
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
	if u, err := url.Parse(path); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		absPath = u.String()
	} else if p, err := filepath.Abs(path); err == nil {
		absPath = p
	}

	if visited[absPath] {
		log.Warn("circular include detected", zap.String("path", absPath))
		return make(map[string]any), nil
	}
	log.Info("reading config", zap.String("path", absPath))
	visited[absPath] = true

	reader, ext, err := readFrom(ctx, path)
	if err != nil {
		return nil, err
	}

	raw, err := parseConfig(ctx, ext, reader, path)
	if err != nil {
		return nil, err
	}
	log.Debug("parsed config", zap.String("path", path))

	raw, err = readIncludedFiles(ctx, includeField, raw, path, visited)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func readIncludedFiles(ctx context.Context, includeField string, raw map[string]any, path string, visited map[string]bool) (map[string]any, error) {
	log := log.FromContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	if includes, ok := raw[includeField].([]any); ok {
		for _, inc := range includes {
			if pattern, ok := inc.(string); ok {
				log.Info("processing include", zap.String("from", path), zap.String("pattern", pattern))
				included, err := mergeFromPattern(ctx, includeField, pattern, visited)
				if err != nil {
					log.Error("pailed to process include", zap.String("pattern", pattern), zap.Error(err))
					return nil, err
				}
				raw, err = deepMerge(included, raw)
				if err != nil {
					log.Error("peep merge failed during include", zap.String("pattern", pattern), zap.Error(err))
					return nil, err
				}
				log.Debug("include merged", zap.String("pattern", pattern))
			}
		}
	}
	return raw, nil
}

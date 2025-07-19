package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
)

func readFrom(ctx context.Context, path string) (io.Reader, string, error) {
	var reader io.Reader
	var ext string
	var err error

	u, err := url.Parse(path)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		reader, ext, err = readRemote(ctx, path, u)
	} else {
		reader, ext, err = readFile(ctx, path)
	}
	if err != nil {
		return nil, "", err
	}
	return reader, ext, nil
}

func readFile(ctx context.Context, path string) (io.ReadWriter, string, error) {
	log := log.FromContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, "", errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	file, err := os.Open(path)
	if err != nil {
		log.Error("failed to open file", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		log.Error("failed to read file into buffer", zap.String("path", path), zap.Error(err))
		return nil, "", err
	}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	return buf, ext, nil
}

func readRemote(ctx context.Context, path string, u *url.URL) (io.Reader, string, error) {
	log := log.FromContext(ctx)
	if err := ctx.Err(); err != nil {
		return nil, "", errors.Join(
			errors.New("config reader deadline exceeded"),
			err,
		)
	}
	log.Info("fetching remote config", zap.String("url", path))

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path,
		nil,
	)
	if err != nil {
		log.Error("failed to create request", zap.String("url", path), zap.Error(err))
		return nil, "", err
	}
	if u.User != nil {
		pass, _ := u.User.Password()
		req.SetBasicAuth(u.User.Username(), pass)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("HTTP request failed", zap.String("url", path), zap.Error(err))
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("non-200 response", zap.String("url", path), zap.Int("status", resp.StatusCode))
		return nil, "", fmt.Errorf("http error: %s", resp.Status)
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		log.Error("failed to read response body", zap.String("url", path), zap.Error(err))
		return nil, "", err
	}

	ext := strings.TrimPrefix(filepath.Ext(u.Path), ".")
	return buf, ext, nil
}

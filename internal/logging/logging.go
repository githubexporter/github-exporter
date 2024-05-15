package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/githubexporter/github-exporter/internal/config"
)

func New(cfg *config.Config, w io.Writer) (*slog.Logger, error) {
	var handler slog.Handler

	var level slog.Level
	err := level.UnmarshalText([]byte(cfg.LogLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to configure logging level: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(cfg.LogFormat) {
	case "plain", "text":
		handler = slog.NewTextHandler(w, opts)
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	default:
		return nil, fmt.Errorf("unknown logging format: %s", cfg.LogFormat)
	}
	return slog.New(handler), nil
}

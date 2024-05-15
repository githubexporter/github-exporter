package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

func New(logLevel string, logFormat string, w io.Writer) (*slog.Logger, error) {
	var handler slog.Handler

	var level slog.Level
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to configure logging level: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch strings.ToLower(logFormat) {
	case "plain", "text":
		handler = slog.NewTextHandler(w, opts)
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	default:
		return nil, fmt.Errorf("unknown logging format: %s", logFormat)
	}
	return slog.New(handler), nil
}

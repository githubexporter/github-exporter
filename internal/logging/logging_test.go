package logging

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	testcases := map[string]struct {
		logLevel             string
		logFormat            string
		expectedErrorMessage string
	}{
		"valid request text": {
			logLevel:             "debug",
			logFormat:            "text",
			expectedErrorMessage: "",
		},
		"valid request plain": {
			logLevel:             "debug",
			logFormat:            "text",
			expectedErrorMessage: "",
		},
		"valid request json": {
			logLevel:             "debug",
			logFormat:            "json",
			expectedErrorMessage: "",
		},
		"invalid log level": {
			logLevel:             "disaster",
			logFormat:            "text",
			expectedErrorMessage: "failed to configure logging level: slog: level string \"disaster\": unknown name",
		},
		"invalid log format": {
			logLevel:             "debug",
			logFormat:            "hieroglyphics",
			expectedErrorMessage: "unknown logging format: hieroglyphics",
		},
	}

	for desc, tc := range testcases {
		t.Run(desc, func(t *testing.T) {
			_, err := New(tc.logLevel, tc.logFormat, nil)
			if tc.expectedErrorMessage == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrorMessage)
			}
		})
	}
}

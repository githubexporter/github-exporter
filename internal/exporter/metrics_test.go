package exporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	metrics := GetMetrics()
	assert.Equal(t, len(metrics), 10)
}

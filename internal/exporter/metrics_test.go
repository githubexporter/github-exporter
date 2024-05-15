package exporter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMetrics(t *testing.T) {
	metrics := GetMetrics()
	assert.Equal(t, len(metrics), 10)
}

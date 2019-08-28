package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewLibvirtExporter(t *testing.T) {
	exporter := NewLibvirtExporter()
	err := prometheus.Register(exporter)
	require.NoError(t, err)
}

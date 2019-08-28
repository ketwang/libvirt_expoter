package exporter

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewLibvirtExporter(t *testing.T) {
	exporter := NewLibvirtExporter()
	err := prometheus.Register(exporter)
	require.NoError(t, err)
}

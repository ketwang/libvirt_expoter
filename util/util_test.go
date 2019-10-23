package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertToBytes(t *testing.T) {
	text := "100663296 KiB"
	_, err := ConvertToBytes(text)
	require.NoError(t, err)
}

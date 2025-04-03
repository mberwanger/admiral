package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTmpLogger(t *testing.T) {
	logger := newTmpLogger()
	require.NotNil(t, logger, "newTmpLogger should return a valid logger, got nil")
}

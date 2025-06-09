package testutil

import (
	"testing"

	"github.com/pt-xyz-multifinance/pkg/logger"
)

// InitTestLogger initializes the logger for tests
func InitTestLogger(t *testing.T) error {
	// Initialize logger with test configuration
	return logger.InitLogger("debug", "console", []string{"stdout"})
}

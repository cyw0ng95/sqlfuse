package common

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestInitLogger(t *testing.T) {
	// Save original logger
	originalLogger := Logger

	// Initialize logger
	InitLogger()

	// Verify logger is initialized
	if Logger.GetLevel() == zerolog.Disabled {
		t.Error("Logger should not be disabled after initialization")
	}

	// Verify we can log without panic
	Logger.Info().Msg("Test message")
	Logger.Error().Msg("Test error")
	Logger.Debug().Msg("Test debug")

	// Restore original logger
	Logger = originalLogger
}

func TestInitLogger_TimeFieldFormat(t *testing.T) {
	// Initialize logger
	InitLogger()

	// Verify time field format is set
	if zerolog.TimeFieldFormat != "01/02 15:04:05" {
		t.Errorf("Expected time format '01/02 15:04:05', got '%s'", zerolog.TimeFieldFormat)
	}
}

func TestInitLogger_NoColor_NotTerminal(t *testing.T) {
	// When stderr is not a terminal, logger should disable colors
	// This test validates that InitLogger doesn't panic when called
	// in a non-terminal environment (like in CI/CD)
	
	// Save original stderr
	originalStderr := os.Stderr
	defer func() { os.Stderr = originalStderr }()

	// Create a pipe (not a terminal)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	os.Stderr = w

	// Should not panic
	InitLogger()
}

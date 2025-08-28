package storage

import (
	"os"
	"testing"
)

func TestGetS3Config_DefaultLogDebug(t *testing.T) {
	// Clear environment variables
	originalS3LogDebug := os.Getenv("S3_LOG_DEBUG")
	os.Unsetenv("S3_LOG_DEBUG")
	defer func() {
		if originalS3LogDebug != "" {
			os.Setenv("S3_LOG_DEBUG", originalS3LogDebug)
		}
	}()

	config, err := getS3Config()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.LogDebug != false {
		t.Errorf("Expected LogDebug to be false by default, got %v", config.LogDebug)
	}
}

func TestGetS3Config_LogDebugTrue(t *testing.T) {
	// Set S3_LOG_DEBUG to "true"
	originalS3LogDebug := os.Getenv("S3_LOG_DEBUG")
	os.Setenv("S3_LOG_DEBUG", "true")
	defer func() {
		if originalS3LogDebug != "" {
			os.Setenv("S3_LOG_DEBUG", originalS3LogDebug)
		} else {
			os.Unsetenv("S3_LOG_DEBUG")
		}
	}()

	config, err := getS3Config()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.LogDebug != true {
		t.Errorf("Expected LogDebug to be true when S3_LOG_DEBUG=true, got %v", config.LogDebug)
	}
}

func TestGetS3Config_LogDebugFalse(t *testing.T) {
	// Set S3_LOG_DEBUG to "false"
	originalS3LogDebug := os.Getenv("S3_LOG_DEBUG")
	os.Setenv("S3_LOG_DEBUG", "false")
	defer func() {
		if originalS3LogDebug != "" {
			os.Setenv("S3_LOG_DEBUG", originalS3LogDebug)
		} else {
			os.Unsetenv("S3_LOG_DEBUG")
		}
	}()

	config, err := getS3Config()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.LogDebug != false {
		t.Errorf("Expected LogDebug to be false when S3_LOG_DEBUG=false, got %v", config.LogDebug)
	}
}

func TestGetS3Config_LogDebugInvalidValue(t *testing.T) {
	// Set S3_LOG_DEBUG to an invalid value
	originalS3LogDebug := os.Getenv("S3_LOG_DEBUG")
	os.Setenv("S3_LOG_DEBUG", "invalid")
	defer func() {
		if originalS3LogDebug != "" {
			os.Setenv("S3_LOG_DEBUG", originalS3LogDebug)
		} else {
			os.Unsetenv("S3_LOG_DEBUG")
		}
	}()

	config, err := getS3Config()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Invalid values should default to false
	if config.LogDebug != false {
		t.Errorf("Expected LogDebug to be false when S3_LOG_DEBUG=invalid, got %v", config.LogDebug)
	}
}
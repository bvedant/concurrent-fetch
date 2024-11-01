package models

import (
	"bytes"
	"errors"
	"testing"
)

func TestResult(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		err         error
		expectData  []byte
		expectError bool
	}{
		{
			name:        "successful result",
			data:        []byte("test data"),
			err:         nil,
			expectData:  []byte("test data"),
			expectError: false,
		},
		{
			name:        "error result",
			data:        nil,
			err:         errors.New("test error"),
			expectData:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Result{
				Data:  tt.data,
				Error: tt.err,
			}

			if tt.expectError && result.Error == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && result.Error != nil {
				t.Errorf("expected no error, got %v", result.Error)
			}
			if !bytes.Equal(result.Data, tt.expectData) {
				t.Errorf("expected data %v, got %v", tt.expectData, result.Data)
			}
		})
	}
}

func TestResultWithNilValues(t *testing.T) {
	result := Result{
		Data:  nil,
		Error: nil,
	}

	if result.Data != nil {
		t.Error("Expected nil data")
	}
	if result.Error != nil {
		t.Error("Expected nil error")
	}
}

func TestResultWithLargeData(t *testing.T) {
	largeData := make([]byte, 1024*1024) // 1MB of data
	result := Result{
		Data: largeData,
	}

	if len(result.Data) != len(largeData) {
		t.Errorf("Expected data length %d, got %d", len(largeData), len(result.Data))
	}
}

func TestResultWithCustomError(t *testing.T) {
	customErr := errors.New("custom error message")
	result := Result{
		Error: customErr,
	}

	if result.Error.Error() != "custom error message" {
		t.Errorf("Expected error message 'custom error message', got '%v'", result.Error)
	}
}

func TestResultWithError(t *testing.T) {
	customErr := errors.New("custom error message")
	result := Result{
		Error: customErr,
	}

	if result.Error != customErr {
		t.Errorf("Expected error %v, got %v", customErr, result.Error)
	}
}

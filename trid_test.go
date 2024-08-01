package trid

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestScan(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		numberOfMatches int
		expectedExt     string
		expectedErr     error
		expectedName    string
	}{
		{
			name:            "Empty filepath",
			numberOfMatches: 1,
			expectedErr:     ErrNoFileSpecified,
		},
		{
			name:            "Non-existent file",
			filePath:        "non_existent_file.txt",
			numberOfMatches: 1,
			expectedErr:     ErrFileNotFound,
		},
		{
			name:            "Invalid number of matches",
			filePath:        "testdata/sample.pdf",
			numberOfMatches: 0,
			expectedErr:     ErrNumberOfMatches,
		},
		{
			name:            "Valid PDF file",
			filePath:        "testdata/sample.pdf",
			numberOfMatches: 1,
			expectedExt:     ".pdf",
			expectedName:    "Adobe Portable Document Format",
		},
		{
			name:            "Valid 7z file",
			filePath:        "testdata/sample.7z",
			numberOfMatches: 1,
			expectedExt:     ".7z",
			expectedName:    "7-Zip compressed archive (v0.4)",
		},
		{
			name:            "Unknown file type",
			filePath:        "testdata/sample.unknown",
			numberOfMatches: 1,
			expectedErr:     ErrUnknownFileType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trid := NewTrid(Options{})
			results, err := trid.Scan(tt.filePath, tt.numberOfMatches)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("Scan() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			} else {
				if len(results) == 0 && tt.expectedExt != "" {
					t.Errorf("Scan() returned no results for %s", tt.filePath)
					return
				}

				if len(results) > 0 && results[0].Extension != tt.expectedExt {
					t.Errorf("Scan() got extension %s, want %s", results[0].Extension, tt.expectedExt)
				}
			}
		})
	}
}

func TestScanErr(t *testing.T) {
	testFile := "./testdata/sample.pdf"

	t.Run("Test non existent command", func(t *testing.T) {
		trid := NewTrid(Options{Cmd: "/unknown-command"})
		_, err := trid.Scan(testFile, 1)
		if err == nil {
			t.Error("Expected an error for non-existent command, but got nil")
		}
	})

	t.Run("Test timeout", func(t *testing.T) {
		trid := NewTrid(Options{Timeout: 1 * time.Millisecond})
		_, err := trid.Scan(testFile, 1)
		if err == nil || !strings.Contains(err.Error(), "command timed out") {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	})

	t.Run("Test empty definitions package", func(t *testing.T) {
		trid := NewTrid(Options{Definitions: "./testdata/empty_def"})
		_, err := trid.Scan(testFile, 1)
		if !errors.Is(err, ErrEmptyDefPackage) {
			t.Errorf("Expected ErrEmptyDefPackage, got: %v", err)
		}
	})
}

package utils

import (
	"strings"
	"testing"
)

// Sample SRT content for testing
const sampleSRT = `1
00:00:04,515 --> 00:00:31,615
<b>Alih Bahasa: CemonK</b>

2
00:00:32,039 --> 00:00:34,050
Tarik napas dalam-dalam.

3
00:00:35,100 --> 00:00:37,200
Dan buang perlahan.
`

// Sample VTT content for testing
const sampleVTT = `WEBVTT

00:00:04.515 --> 00:00:31.615
<b>Alih Bahasa: CemonK</b>

00:00:32.039 --> 00:00:34.050
Tarik napas dalam-dalam.

00:00:35.100 --> 00:00:37.200
Dan buang perlahan.
`

func TestDetectFormat(t *testing.T) {
	converter := NewSubtitleConverter()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Detect SRT format",
			content:  sampleSRT,
			expected: "srt",
		},
		{
			name:     "Detect VTT format",
			content:  sampleVTT,
			expected: "vtt",
		},
		{
			name:     "Empty content",
			content:  "",
			expected: "unknown",
		},
		{
			name:     "Invalid format",
			content:  "Just some random text\nwithout any subtitle markers",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.DetectFormat(tt.content)
			if result != tt.expected {
				t.Errorf("DetectFormat() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSRTToVTT(t *testing.T) {
	converter := NewSubtitleConverter()

	tests := []struct {
		name        string
		input       string
		expectError bool
		checkOutput func(string) bool
	}{
		{
			name:        "Convert valid SRT to VTT",
			input:       sampleSRT,
			expectError: false,
			checkOutput: func(output string) bool {
				// Should start with WEBVTT
				if !strings.HasPrefix(output, "WEBVTT") {
					return false
				}
				// Should not contain sequence numbers
				if strings.Contains(output, "\n1\n") || strings.Contains(output, "\n2\n") {
					return false
				}
				// Should use dots instead of commas in timestamps
				if strings.Contains(output, "00:00:04,515") {
					return false
				}
				if !strings.Contains(output, "00:00:04.515") {
					return false
				}
				return true
			},
		},
		{
			name:        "Empty content",
			input:       "",
			expectError: true,
			checkOutput: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.SRTToVTT(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkOutput != nil && !tt.checkOutput(result) {
				t.Errorf("Output validation failed.\nGot:\n%s", result)
			}
		})
	}
}

func TestVTTToSRT(t *testing.T) {
	converter := NewSubtitleConverter()

	tests := []struct {
		name        string
		input       string
		expectError bool
		checkOutput func(string) bool
	}{
		{
			name:        "Convert valid VTT to SRT",
			input:       sampleVTT,
			expectError: false,
			checkOutput: func(output string) bool {
				// Should not contain WEBVTT header
				if strings.HasPrefix(output, "WEBVTT") {
					return false
				}
				// Should contain sequence numbers
				if !strings.Contains(output, "1\n") {
					return false
				}
				// Should use commas instead of dots in timestamps
				if strings.Contains(output, "00:00:04.515") {
					return false
				}
				if !strings.Contains(output, "00:00:04,515") {
					return false
				}
				return true
			},
		},
		{
			name:        "Empty content",
			input:       "",
			expectError: true,
			checkOutput: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.VTTToSRT(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkOutput != nil && !tt.checkOutput(result) {
				t.Errorf("Output validation failed.\nGot:\n%s", result)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	converter := NewSubtitleConverter()

	tests := []struct {
		name        string
		content     string
		fromFormat  string
		toFormat    string
		expectError bool
	}{
		{
			name:        "SRT to VTT",
			content:     sampleSRT,
			fromFormat:  "srt",
			toFormat:    "vtt",
			expectError: false,
		},
		{
			name:        "VTT to SRT",
			content:     sampleVTT,
			fromFormat:  "vtt",
			toFormat:    "srt",
			expectError: false,
		},
		{
			name:        "Same format (no conversion)",
			content:     sampleSRT,
			fromFormat:  "srt",
			toFormat:    "srt",
			expectError: false,
		},
		{
			name:        "Unsupported conversion",
			content:     sampleSRT,
			fromFormat:  "srt",
			toFormat:    "ass",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.content, tt.fromFormat, tt.toFormat)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == "" {
				t.Error("Expected non-empty result")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	converter := NewSubtitleConverter()

	tests := []struct {
		name        string
		content     string
		format      string
		expectError bool
	}{
		{
			name:        "Valid SRT",
			content:     sampleSRT,
			format:      "srt",
			expectError: false,
		},
		{
			name:        "Valid VTT",
			content:     sampleVTT,
			format:      "vtt",
			expectError: false,
		},
		{
			name:        "Empty content",
			content:     "",
			format:      "srt",
			expectError: true,
		},
		{
			name:        "Format mismatch",
			content:     sampleSRT,
			format:      "vtt",
			expectError: true,
		},
		{
			name:        "Unknown format",
			content:     "random text",
			format:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := converter.Validate(tt.content, tt.format)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test round-trip conversion (SRT -> VTT -> SRT)
func TestRoundTripConversion(t *testing.T) {
	converter := NewSubtitleConverter()

	// Convert SRT to VTT
	vtt, err := converter.SRTToVTT(sampleSRT)
	if err != nil {
		t.Fatalf("SRT to VTT conversion failed: %v", err)
	}

	// Convert back to SRT
	srt, err := converter.VTTToSRT(vtt)
	if err != nil {
		t.Fatalf("VTT to SRT conversion failed: %v", err)
	}

	// Check that essential content is preserved
	if !strings.Contains(srt, "Alih Bahasa: CemonK") {
		t.Error("Content lost during round-trip conversion")
	}

	if !strings.Contains(srt, "Tarik napas dalam-dalam") {
		t.Error("Content lost during round-trip conversion")
	}

	// Check timestamp format in final SRT
	if !strings.Contains(srt, "00:00:04,515") {
		t.Error("Timestamp format incorrect after round-trip")
	}
}

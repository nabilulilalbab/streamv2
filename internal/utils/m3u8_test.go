package utils

import (
	"testing"
)

// TestConvertMP4ToM3U8 tests the ConvertMP4ToM3U8 function with various inputs
func TestConvertMP4ToM3U8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Convert .txt to .m3u8",
			input:    "https://jeniusplay.com/cdn/hls/4fa8e315e2d598d24af31cd662d0e5ec/master.txt",
			expected: "https://jeniusplay.com/cdn/hls/4fa8e315e2d598d24af31cd662d0e5ec/master.m3u8",
		},
		{
			name:     "Convert .mp4 to .m3u8",
			input:    "https://example.com/videos/movie.mp4",
			expected: "https://example.com/videos/movie.m3u8",
		},
		{
			name:     "Convert .avi to .m3u8",
			input:    "https://example.com/videos/movie.avi",
			expected: "https://example.com/videos/movie.m3u8",
		},
		{
			name:     "Convert .mkv to .m3u8",
			input:    "https://example.com/videos/movie.mkv",
			expected: "https://example.com/videos/movie.m3u8",
		},
		{
			name:     "No extension - append .m3u8",
			input:    "https://example.com/videos/movie",
			expected: "https://example.com/videos/movie.m3u8",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "URL with dots in path but no file extension",
			input:    "https://example.com/v1.0/videos/movie",
			expected: "https://example.com/v1.0/videos/movie.m3u8",
		},
		{
			name:     "URL with dots in path and file extension",
			input:    "https://example.com/v1.0/videos/movie.mp4",
			expected: "https://example.com/v1.0/videos/movie.m3u8",
		},
		{
			name:     "Already .m3u8 extension",
			input:    "https://example.com/videos/movie.m3u8",
			expected: "https://example.com/videos/movie.m3u8",
		},
		{
			name:     "Simple filename with .txt",
			input:    "master.txt",
			expected: "master.m3u8",
		},
		{
			name:     "Simple filename with .mp4",
			input:    "video.mp4",
			expected: "video.m3u8",
		},
		{
			name:     "Complex path with query parameters (no extension in filename)",
			input:    "https://example.com/videos/stream?id=123",
			expected: "https://example.com/videos/stream.m3u8?id=123",
		},
		{
			name:     "Path with multiple dots",
			input:    "https://example.com/path/file.name.with.dots.txt",
			expected: "https://example.com/path/file.name.with.dots.m3u8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMP4ToM3U8(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertMP4ToM3U8(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestConvertMP4ToM3U8_PythonParity tests that the Go implementation matches Python behavior
func TestConvertMP4ToM3U8_PythonParity(t *testing.T) {
	// This is the exact case from the production bug
	input := "https://jeniusplay.com/cdn/hls/4fa8e315e2d598d24af31cd662d0e5ec/master.txt"
	expected := "https://jeniusplay.com/cdn/hls/4fa8e315e2d598d24af31cd662d0e5ec/master.m3u8"
	
	result := ConvertMP4ToM3U8(input)
	
	if result != expected {
		t.Errorf("Production bug case failed!\nInput:    %q\nExpected: %q\nGot:      %q", input, expected, result)
	}
}

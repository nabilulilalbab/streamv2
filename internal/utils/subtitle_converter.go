package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// SubtitleConverter handles subtitle format conversion
type SubtitleConverter struct{}

// NewSubtitleConverter creates a new subtitle converter
func NewSubtitleConverter() *SubtitleConverter {
	return &SubtitleConverter{}
}

// Convert converts subtitle between formats
func (c *SubtitleConverter) Convert(content string, fromFormat, toFormat string) (string, error) {
	// Normalize format strings
	fromFormat = strings.ToLower(strings.TrimSpace(fromFormat))
	toFormat = strings.ToLower(strings.TrimSpace(toFormat))

	// If same format, return as-is
	if fromFormat == toFormat {
		return content, nil
	}

	// Convert based on format pair
	switch {
	case fromFormat == "srt" && toFormat == "vtt":
		return c.SRTToVTT(content)
	case fromFormat == "vtt" && toFormat == "srt":
		return c.VTTToSRT(content)
	default:
		return "", fmt.Errorf("unsupported conversion: %s to %s", fromFormat, toFormat)
	}
}

// SRTToVTT converts SRT subtitle to VTT format
func (c *SubtitleConverter) SRTToVTT(srtContent string) (string, error) {
	if strings.TrimSpace(srtContent) == "" {
		return "", fmt.Errorf("empty subtitle content")
	}

	var result strings.Builder
	
	// Add WEBVTT header
	result.WriteString("WEBVTT\n\n")

	// Split into lines
	lines := strings.Split(srtContent, "\n")
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Remove BOM if present in this line
		trimmed = strings.TrimPrefix(trimmed, "\uFEFF")
		
		// Skip sequence numbers (lines with only digits)
		if matched, _ := regexp.MatchString(`^\d+$`, trimmed); matched {
			continue
		}
		
		// Convert timestamp format: replace comma with dot
		// SRT: 00:00:04,515 --> 00:00:31,615
		// VTT: 00:00:04.515 --> 00:00:31.615
		if strings.Contains(trimmed, " --> ") {
			converted := strings.ReplaceAll(trimmed, ",", ".")
			result.WriteString(converted + "\n")
		} else {
			// Keep other lines as-is (cue text, blank lines)
			result.WriteString(line + "\n")
		}
	}

	return result.String(), nil
}

// VTTToSRT converts VTT subtitle to SRT format
func (c *SubtitleConverter) VTTToSRT(vttContent string) (string, error) {
	if strings.TrimSpace(vttContent) == "" {
		return "", fmt.Errorf("empty subtitle content")
	}

	var result strings.Builder
	
	// Split into lines
	lines := strings.Split(vttContent, "\n")
	
	// Track sequence number
	sequenceNum := 0
	inCue := false
	skipNext := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip WEBVTT header
		if strings.HasPrefix(trimmed, "WEBVTT") {
			continue
		}
		
		// Skip cue identifiers (optional identifiers before timestamps)
		if skipNext {
			skipNext = false
			continue
		}
		
		// Detect timestamp line
		if strings.Contains(trimmed, " --> ") {
			sequenceNum++
			
			// Add sequence number
			result.WriteString(fmt.Sprintf("%d\n", sequenceNum))
			
			// Convert timestamp format: replace dot with comma
			// VTT: 00:00:04.515 --> 00:00:31.615
			// SRT: 00:00:04,515 --> 00:00:31,615
			converted := strings.ReplaceAll(trimmed, ".", ",")
			// But don't replace dots in actual numbers (keep HH:MM:SS format)
			// Only replace the millisecond separator
			converted = regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})`).ReplaceAllString(
				strings.ReplaceAll(trimmed, ".", ","),
				"$1,$2",
			)
			result.WriteString(converted + "\n")
			inCue = true
		} else if trimmed == "" {
			// Blank line - end of cue
			result.WriteString("\n")
			inCue = false
		} else if inCue {
			// Cue text
			result.WriteString(line + "\n")
		}
	}

	return result.String(), nil
}

// DetectFormat detects subtitle format from content
func (c *SubtitleConverter) DetectFormat(content string) string {
	trimmed := strings.TrimSpace(content)
	
	if trimmed == "" {
		return "unknown"
	}

	// Remove BOM if present
	trimmed = strings.TrimPrefix(trimmed, "\uFEFF")

	// Check for WEBVTT header (VTT format)
	if strings.HasPrefix(trimmed, "WEBVTT") {
		return "vtt"
	}

	// Check for SRT format characteristics
	// SRT typically starts with a sequence number followed by timestamp
	lines := strings.Split(trimmed, "\n")
	if len(lines) >= 2 {
		// First line should be a number (sequence number)
		firstLine := strings.TrimSpace(lines[0])
		// Remove BOM from first line if present
		firstLine = strings.TrimPrefix(firstLine, "\uFEFF")
		
		if matched, _ := regexp.MatchString(`^\d+$`, firstLine); matched {
			// Second line should contain timestamp
			secondLine := strings.TrimSpace(lines[1])
			if strings.Contains(secondLine, " --> ") {
				// Check if it uses comma (SRT) or dot (VTT) for milliseconds
				if strings.Contains(secondLine, ",") {
					return "srt"
				}
			}
		}
	}

	// Check for VTT without header (some files might be malformed)
	// VTT uses dots for milliseconds, SRT uses commas
	if strings.Contains(trimmed, " --> ") {
		// Count commas vs dots in timestamp lines
		commaCount := 0
		dotCount := 0
		
		timestampRe := regexp.MustCompile(`\d{2}:\d{2}:\d{2}[.,]\d{3} --> \d{2}:\d{2}:\d{2}[.,]\d{3}`)
		timestamps := timestampRe.FindAllString(trimmed, -1)
		
		for _, ts := range timestamps {
			if strings.Contains(ts, ",") {
				commaCount++
			}
			if strings.Count(ts, ".") >= 2 { // At least 2 dots for milliseconds
				dotCount++
			}
		}
		
		if commaCount > dotCount {
			return "srt"
		} else if dotCount > 0 {
			return "vtt"
		}
	}

	return "unknown"
}

// Validate validates subtitle format
func (c *SubtitleConverter) Validate(content string, format string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("empty subtitle content")
	}

	detected := c.DetectFormat(content)
	
	if detected == "unknown" {
		return fmt.Errorf("unable to detect subtitle format")
	}

	if format != "" && detected != strings.ToLower(format) {
		return fmt.Errorf("content appears to be %s format, not %s", detected, format)
	}

	return nil
}

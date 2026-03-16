package utils

import (
	"fmt"
	"io"
	"strings"

	"idlix-api/internal/models"

	"github.com/grafov/m3u8"
)

// M3U8Parser handles M3U8 playlist parsing
type M3U8Parser struct {
	client *HTTPClient
}

// NewM3U8Parser creates a new M3U8 parser
func NewM3U8Parser(client *HTTPClient) *M3U8Parser {
	return &M3U8Parser{
		client: client,
	}
}

// ConvertMP4ToM3U8 converts MP4 URL to M3U8 URL
func ConvertMP4ToM3U8(mp4URL string) string {
	// Replace .mp4 extension with .m3u8
	if strings.HasSuffix(mp4URL, ".mp4") {
		return strings.TrimSuffix(mp4URL, ".mp4") + ".m3u8"
	}
	return mp4URL
}

// ParseMasterPlaylist parses M3U8 master playlist and extracts variants
func (p *M3U8Parser) ParseMasterPlaylist(m3u8URL string) ([]models.VariantPlaylist, bool, error) {
	// Fetch M3U8 playlist
	resp, err := p.client.Get(m3u8URL, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch M3U8 playlist: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != 200 {
		return nil, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read M3U8 response: %w", err)
	}

	// Parse M3U8 playlist
	playlist, listType, err := m3u8.DecodeFrom(strings.NewReader(string(body)), true)
	if err != nil {
		return nil, false, fmt.Errorf("failed to decode M3U8 playlist: %w", err)
	}

	// Check if it's a master playlist (variant playlist)
	if listType == m3u8.MASTER {
		masterPlaylist := playlist.(*m3u8.MasterPlaylist)
		
		if len(masterPlaylist.Variants) == 0 {
			return nil, false, fmt.Errorf("no variants found in master playlist")
		}

		// Extract variants
		var variants []models.VariantPlaylist
		for i, variant := range masterPlaylist.Variants {
			if variant == nil {
				continue
			}

			// Build variant info
			variantInfo := models.VariantPlaylist{
				ID:        fmt.Sprintf("%d", i),
				Bandwidth: variant.Bandwidth,
				URI:       variant.URI,
			}

			// Add resolution if available
			if variant.Resolution != "" {
				variantInfo.Resolution = variant.Resolution
			}

			// Make URI absolute if it's relative
			if !strings.HasPrefix(variant.URI, "http") {
				// Extract base URL from m3u8URL (protocol + host)
				if strings.Contains(m3u8URL, "://") {
					parts := strings.SplitN(m3u8URL, "/", 4) // ["https:", "", "jeniusplay.com", "path..."]
					if len(parts) >= 3 {
						baseURL := parts[0] + "//" + parts[1] + "/" + parts[2]
						// Clean up variant URI (remove leading slash if exists)
						cleanURI := strings.TrimPrefix(variant.URI, "/")
						variantInfo.URI = baseURL + "/" + cleanURI
					}
				}
			}

			variants = append(variants, variantInfo)
		}

		// Sort variants by bandwidth (highest first)
		sortVariantsByBandwidth(variants)

		return variants, true, nil
	}

	// If it's a media playlist (single variant), return as single item
	if listType == m3u8.MEDIA {
		variant := models.VariantPlaylist{
			ID:        "0",
			Bandwidth: 0,
			URI:       m3u8URL,
			Resolution: "unknown",
		}
		return []models.VariantPlaylist{variant}, false, nil
	}

	return nil, false, fmt.Errorf("unknown playlist type")
}

// sortVariantsByBandwidth sorts variants by bandwidth (highest first)
func sortVariantsByBandwidth(variants []models.VariantPlaylist) {
	// Simple bubble sort (efficient for small arrays)
	n := len(variants)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if variants[j].Bandwidth < variants[j+1].Bandwidth {
				variants[j], variants[j+1] = variants[j+1], variants[j]
			}
		}
	}
}

// GetHighestQuality returns the variant with highest bandwidth
func GetHighestQuality(variants []models.VariantPlaylist) *models.VariantPlaylist {
	if len(variants) == 0 {
		return nil
	}

	highest := &variants[0]
	for i := range variants {
		if variants[i].Bandwidth > highest.Bandwidth {
			highest = &variants[i]
		}
	}

	return highest
}

// FindVariantByResolution finds variant by resolution string (e.g., "1920x1080")
func FindVariantByResolution(variants []models.VariantPlaylist, resolution string) *models.VariantPlaylist {
	for i := range variants {
		if variants[i].Resolution == resolution {
			return &variants[i]
		}
	}
	return nil
}

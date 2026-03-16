package repositories

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"idlix-api/internal/models"
	"idlix-api/internal/utils"
)

// JeniusRepository handles JeniusPlay API interactions
type JeniusRepository struct {
	client  *utils.HTTPClient
	baseURL string
}

// NewJeniusRepository creates a new JeniusPlay repository
func NewJeniusRepository(client *utils.HTTPClient, baseURL string) *JeniusRepository {
	return &JeniusRepository{
		client:  client,
		baseURL: baseURL,
	}
}

// GetVideoSource gets video source URL from JeniusPlay
func (r *JeniusRepository) GetVideoSource(embedHash string) (*models.JeniusPlayVideoResponse, error) {
	if embedHash == "" {
		return nil, fmt.Errorf("embed hash is required")
	}

	// Prepare form data
	formData := url.Values{}
	formData.Set("hash", embedHash)
	formData.Set("r", "https://tv12.idlixku.com/")

	// Request headers
	headers := map[string]string{
		"Host":         "jeniusplay.com",
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"X-Requested-With": "XMLHttpRequest",
	}

	// POST request
	resp, err := r.client.Post(
		r.baseURL+"player/index.php?data="+embedHash+"&do=getVideo",
		headers,
		formData.Encode(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video source: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var videoResp models.JeniusPlayVideoResponse
	if err := json.Unmarshal(body, &videoResp); err != nil {
		return nil, fmt.Errorf("failed to parse video response: %w", err)
	}

	// Validate response
	if videoResp.VideoSource == "" {
		return nil, fmt.Errorf("video source not found in response")
	}

	return &videoResp, nil
}

// GetSubtitleURL extracts subtitle URL from JeniusPlay response
func (r *JeniusRepository) GetSubtitleURL(embedHash string) (string, error) {
	if embedHash == "" {
		return "", fmt.Errorf("embed hash is required")
	}

	// Prepare form data
	formData := url.Values{}
	formData.Set("hash", embedHash)
	formData.Set("r", "https://tv12.idlixku.com/")

	// Request headers
	headers := map[string]string{
		"Host":         "jeniusplay.com",
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	// POST request
	resp, err := r.client.Post(
		r.baseURL+"player/index.php?data="+embedHash+"&do=getVideo",
		headers,
		formData.Encode(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subtitle: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Extract subtitle using regex
	// Pattern: var playerjsSubtitle = "...";
	re := regexp.MustCompile(`var playerjsSubtitle = "(.*)";`)
	matches := re.FindStringSubmatch(string(body))

	if len(matches) < 2 {
		return "", fmt.Errorf("subtitle not found")
	}

	subtitleURL := matches[1]
	
	// Extract https URL from the match
	if strings.Contains(subtitleURL, "https://") {
		parts := strings.Split(subtitleURL, "https://")
		if len(parts) > 1 {
			return "https://" + parts[1], nil
		}
	}

	return "", fmt.Errorf("invalid subtitle URL format")
}

// ExtractEmbedHash extracts hash from embed URL
func (r *JeniusRepository) ExtractEmbedHash(embedURL string) (string, error) {
	if embedURL == "" {
		return "", fmt.Errorf("embed URL is required")
	}

	parsedURL, err := url.Parse(embedURL)
	if err != nil {
		return "", fmt.Errorf("invalid embed URL: %w", err)
	}

	// Check for /video/ in path
	if strings.Contains(parsedURL.Path, "/video/") {
		parts := strings.Split(parsedURL.Path, "/")
		for i, part := range parts {
			if part == "video" && i+1 < len(parts) {
				return parts[i+1], nil
			}
		}
	}

	// Check for query parameter
	query := parsedURL.Query()
	if hash := query.Get("data"); hash != "" {
		return hash, nil
	}
	if hash := query.Get("hash"); hash != "" {
		return hash, nil
	}

	return "", fmt.Errorf("could not extract hash from embed URL")
}

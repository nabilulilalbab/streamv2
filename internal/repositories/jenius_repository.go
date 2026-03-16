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
	fmt.Printf("\n🔍 [JENIUS_REPO] GetVideoSource called with embedHash: %s\n", embedHash)
	
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

	requestURL := r.baseURL + "player/index.php?data=" + embedHash + "&do=getVideo"
	fmt.Printf("📤 [JENIUS_REPO] POST Request to: %s\n", requestURL)
	fmt.Printf("📤 [JENIUS_REPO] Headers: %+v\n", headers)
	fmt.Printf("📤 [JENIUS_REPO] Form Data: %s\n", formData.Encode())

	// POST request
	resp, err := r.client.Post(requestURL, headers, formData.Encode())
	if err != nil {
		fmt.Printf("❌ [JENIUS_REPO] POST request failed: %v\n", err)
		return nil, fmt.Errorf("failed to fetch video source: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📥 [JENIUS_REPO] Response Status Code: %d\n", resp.StatusCode)

	// Check status
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ [JENIUS_REPO] Non-200 status code: %d\n", resp.StatusCode)
		fmt.Printf("📥 [JENIUS_REPO] Response Body: %s\n", string(body))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [JENIUS_REPO] Failed to read response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	bodyPreview := string(body)
	if len(body) > 500 {
		bodyPreview = bodyPreview[:500]
	}
	fmt.Printf("📥 [JENIUS_REPO] Response Body (first 500 chars): %s\n", bodyPreview)

	// Parse JSON response
	var videoResp models.JeniusPlayVideoResponse
	if err := json.Unmarshal(body, &videoResp); err != nil {
		fmt.Printf("❌ [JENIUS_REPO] Failed to parse JSON: %v\n", err)
		fmt.Printf("📥 [JENIUS_REPO] Full Response Body: %s\n", string(body))
		return nil, fmt.Errorf("failed to parse video response: %w", err)
	}

	// Validate response
	if videoResp.VideoSource == "" {
		fmt.Printf("❌ [JENIUS_REPO] VideoSource is empty in response\n")
		fmt.Printf("📥 [JENIUS_REPO] Parsed Response: %+v\n", videoResp)
		return nil, fmt.Errorf("video source not found in response")
	}

	fmt.Printf("✅ [JENIUS_REPO] Successfully got VideoSource: %s\n", videoResp.VideoSource)
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
	fmt.Printf("\n🔑 [JENIUS_REPO] ExtractEmbedHash called with: %s\n", embedURL)
	
	if embedURL == "" {
		return "", fmt.Errorf("embed URL is required")
	}

	parsedURL, err := url.Parse(embedURL)
	if err != nil {
		fmt.Printf("❌ [JENIUS_REPO] Failed to parse URL: %v\n", err)
		return "", fmt.Errorf("invalid embed URL: %w", err)
	}

	fmt.Printf("🔍 [JENIUS_REPO] Parsed URL - Scheme: %s, Host: %s, Path: %s, Query: %s\n", 
		parsedURL.Scheme, parsedURL.Host, parsedURL.Path, parsedURL.RawQuery)

	// Check for /video/ in path
	if strings.Contains(parsedURL.Path, "/video/") {
		parts := strings.Split(parsedURL.Path, "/")
		fmt.Printf("🔍 [JENIUS_REPO] Found /video/ in path, parts: %v\n", parts)
		for i, part := range parts {
			if part == "video" && i+1 < len(parts) {
				hash := parts[i+1]
				fmt.Printf("✅ [JENIUS_REPO] Extracted hash from path: %s\n", hash)
				return hash, nil
			}
		}
	}

	// Check for query parameter
	query := parsedURL.Query()
	if hash := query.Get("data"); hash != "" {
		fmt.Printf("✅ [JENIUS_REPO] Extracted hash from 'data' query param: %s\n", hash)
		return hash, nil
	}
	if hash := query.Get("hash"); hash != "" {
		fmt.Printf("✅ [JENIUS_REPO] Extracted hash from 'hash' query param: %s\n", hash)
		return hash, nil
	}

	fmt.Printf("❌ [JENIUS_REPO] Could not extract hash from URL\n")
	return "", fmt.Errorf("could not extract hash from embed URL")
}

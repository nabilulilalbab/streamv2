package repositories

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"idlix-api/internal/models"
	"idlix-api/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

// IDLIXRepository handles scraping IDLIX website
type IDLIXRepository struct {
	client *utils.HTTPClient
}

// NewIDLIXRepository creates a new IDLIX repository
func NewIDLIXRepository(client *utils.HTTPClient) *IDLIXRepository {
	return &IDLIXRepository{
		client: client,
	}
}

// GetFeaturedMovies scrapes featured movies from homepage
func (r *IDLIXRepository) GetFeaturedMovies() ([]models.Movie, error) {
	// Request homepage
	resp, err := r.client.Get(r.client.GetBaseURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch homepage: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract featured movies
	var movies []models.Movie

	// Find featured section: div.items.featured article
	doc.Find("div.items.featured article").Each(func(i int, s *goquery.Selection) {
		// Extract data
		url, urlExists := s.Find("a").Attr("href")
		title := strings.TrimSpace(s.Find("h3").Text())
		poster, _ := s.Find("img").Attr("src")
		
		// Extract year from span (first span element)
		year := strings.TrimSpace(s.Find("span").First().Text())
		
		// Extract type from URL path
		// URL format: https://tv12.idlixku.com/{type}/{slug}/
		// Split and get index 3 (0: https:, 1: empty, 2: tv12.idlixku.com, 3: type)
		urlParts := strings.Split(url, "/")
		movieType := ""
		if len(urlParts) > 3 {
			movieType = urlParts[3]
		}

		// Validate required fields
		if !urlExists || url == "" || title == "" {
			return // Skip invalid entries
		}

		// Filter out TV series (only movies) - match Python logic
		if movieType == "tvseries" {
			return
		}

		// Add to list
		movies = append(movies, models.Movie{
			URL:    url,
			Title:  title,
			Year:   year,
			Type:   movieType,
			Poster: poster,
		})
	})

	if len(movies) == 0 {
		return nil, fmt.Errorf("no featured movies found")
	}

	return movies, nil
}

// GetVideoData extracts video metadata from movie page
func (r *IDLIXRepository) GetVideoData(movieURL string) (videoID, videoName, poster string, err error) {
	// Request movie page
	resp, err := r.client.Get(movieURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to fetch movie page: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != 200 {
		return "", "", "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract video ID from meta tag with id=dooplay-ajax-counter and data-postid attribute
	videoID, exists := doc.Find("meta#dooplay-ajax-counter").Attr("data-postid")
	if !exists || videoID == "" {
		return "", "", "", fmt.Errorf("video ID not found")
	}

	// Extract video name from meta itemprop=name
	videoName, exists = doc.Find("meta[itemprop='name']").Attr("content")
	if !exists || videoName == "" {
		// Fallback to h1.title
		videoName = strings.TrimSpace(doc.Find("h1.title").Text())
		if videoName == "" {
			return "", "", "", fmt.Errorf("video name not found")
		}
	}

	// Extract poster from img itemprop=image
	poster, _ = doc.Find("img[itemprop='image']").Attr("src")
	if poster == "" {
		// Fallback to div.poster img
		poster, _ = doc.Find("div.poster img").Attr("src")
	}

	return videoID, videoName, poster, nil
}

// GetEmbedURL gets and decrypts the embed URL
func (r *IDLIXRepository) GetEmbedURL(videoID string) (string, error) {
	fmt.Printf("\n🔐 [IDLIX_REPO] GetEmbedURL called with videoID: %s\n", videoID)
	
	if videoID == "" {
		return "", fmt.Errorf("video ID is required")
	}

	// Prepare form data
	formData := url.Values{}
	formData.Set("action", "doo_player_ajax")
	formData.Set("post", videoID)
	formData.Set("nume", "1")
	formData.Set("type", "movie")

	// Request headers
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"Referer":      r.client.GetBaseURL(),
		"Origin":       strings.TrimSuffix(r.client.GetBaseURL(), "/"),
	}

	requestURL := r.client.GetBaseURL() + "wp-admin/admin-ajax.php"
	fmt.Printf("📤 [IDLIX_REPO] POST Request to: %s\n", requestURL)
	fmt.Printf("📤 [IDLIX_REPO] Form Data: %s\n", formData.Encode())

	// POST request
	resp, err := r.client.Post(requestURL, headers, formData.Encode())
	if err != nil {
		fmt.Printf("❌ [IDLIX_REPO] POST request failed: %v\n", err)
		return "", fmt.Errorf("failed to fetch embed URL: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📥 [IDLIX_REPO] Response Status Code: %d\n", resp.StatusCode)

	// Check status
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ [IDLIX_REPO] Non-200 status: %d, Body: %s\n", resp.StatusCode, string(body))
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [IDLIX_REPO] Failed to read response: %v\n", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	bodyPreview := string(body)
	if len(body) > 300 {
		bodyPreview = bodyPreview[:300]
	}
	fmt.Printf("📥 [IDLIX_REPO] Response Body (first 300 chars): %s\n", bodyPreview)

	var embedResponse models.EncryptedEmbed
	if err := json.Unmarshal(body, &embedResponse); err != nil {
		fmt.Printf("❌ [IDLIX_REPO] Failed to parse JSON: %v\n", err)
		return "", fmt.Errorf("failed to parse embed response: %w", err)
	}

	// Validate response
	if embedResponse.EmbedURL == "" {
		fmt.Printf("❌ [IDLIX_REPO] EmbedURL is empty in response\n")
		return "", fmt.Errorf("embed URL not found in response")
	}

	embedPreview := embedResponse.EmbedURL
	if len(embedResponse.EmbedURL) > 100 {
		embedPreview = embedResponse.EmbedURL[:100]
	}
	fmt.Printf("🔐 [IDLIX_REPO] Encrypted embed_url: %s\n", embedPreview)
	fmt.Printf("🔑 [IDLIX_REPO] Key: %s\n", embedResponse.Key)

	// Parse the embed_url JSON string to get encrypted data
	var encryptedData models.EncryptedData
	if err := json.Unmarshal([]byte(embedResponse.EmbedURL), &encryptedData); err != nil {
		fmt.Printf("❌ [IDLIX_REPO] Failed to parse encrypted data: %v\n", err)
		return "", fmt.Errorf("failed to parse encrypted embed data: %w", err)
	}

	// Validate encrypted data
	if encryptedData.M == "" {
		fmt.Printf("❌ [IDLIX_REPO] M parameter is empty\n")
		return "", fmt.Errorf("M parameter not found in encrypted data")
	}

	fmt.Printf("🔐 [IDLIX_REPO] Encrypted M parameter: %s\n", encryptedData.M)

	// Generate passphrase using dec()
	passphrase, err := utils.Dec(embedResponse.Key, encryptedData.M)
	if err != nil {
		fmt.Printf("❌ [IDLIX_REPO] Failed to generate passphrase: %v\n", err)
		return "", fmt.Errorf("failed to generate passphrase: %w", err)
	}

	fmt.Printf("🔑 [IDLIX_REPO] Generated passphrase: %s\n", passphrase)

	// Decrypt embed URL
	decryptedJSON, err := utils.CryptoJSDecrypt(embedResponse.EmbedURL, passphrase)
	if err != nil {
		fmt.Printf("❌ [IDLIX_REPO] Decryption failed: %v\n", err)
		return "", fmt.Errorf("failed to decrypt embed URL: %w", err)
	}

	fmt.Printf("✅ [IDLIX_REPO] Decrypted result: %s\n", decryptedJSON)

	// Parse decrypted JSON to extract actual URL
	// The decrypted result is a JSON string containing the actual embed URL
	var embedURLData interface{}
	if err := json.Unmarshal([]byte(decryptedJSON), &embedURLData); err != nil {
		// If it's not JSON, return as-is (might be plain string)
		result := strings.Trim(decryptedJSON, "\"")
		fmt.Printf("✅ [IDLIX_REPO] Final embed URL (plain): %s\n", result)
		return result, nil
	}

	// If it's a string, return it
	if urlStr, ok := embedURLData.(string); ok {
		fmt.Printf("✅ [IDLIX_REPO] Final embed URL (string): %s\n", urlStr)
		return urlStr, nil
	}

	// If it's an object with "embed_url" field
	if embedMap, ok := embedURLData.(map[string]interface{}); ok {
		if url, exists := embedMap["embed_url"]; exists {
			if urlStr, ok := url.(string); ok {
				fmt.Printf("✅ [IDLIX_REPO] Final embed URL (object): %s\n", urlStr)
				return urlStr, nil
			}
		}
	}

	fmt.Printf("✅ [IDLIX_REPO] Final embed URL (raw): %s\n", decryptedJSON)
	return decryptedJSON, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

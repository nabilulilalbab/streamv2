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
		year := strings.TrimSpace(s.Find("span.year").Text())
		movieType := strings.TrimSpace(s.Find("span.type").Text())

		// Validate required fields
		if !urlExists || url == "" || title == "" {
			return // Skip invalid entries
		}

		// Filter out TV series (only movies)
		if strings.ToLower(movieType) == "tv" || strings.Contains(strings.ToLower(title), "season") {
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

	// POST request
	resp, err := r.client.Post(
		r.client.GetBaseURL()+"wp-admin/admin-ajax.php",
		headers,
		formData.Encode(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fetch embed URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var embedResponse models.EncryptedEmbed
	if err := json.Unmarshal(body, &embedResponse); err != nil {
		return "", fmt.Errorf("failed to parse embed response: %w", err)
	}

	// Validate response
	if embedResponse.EmbedURL == "" {
		return "", fmt.Errorf("embed URL not found in response")
	}

	// Parse the embed_url JSON string to get encrypted data
	var encryptedData models.EncryptedData
	if err := json.Unmarshal([]byte(embedResponse.EmbedURL), &encryptedData); err != nil {
		return "", fmt.Errorf("failed to parse encrypted embed data: %w", err)
	}

	// Validate encrypted data
	if encryptedData.M == "" {
		return "", fmt.Errorf("M parameter not found in encrypted data")
	}

	// Generate passphrase using dec()
	passphrase, err := utils.Dec(embedResponse.Key, encryptedData.M)
	if err != nil {
		return "", fmt.Errorf("failed to generate passphrase: %w", err)
	}

	// Decrypt embed URL
	decryptedJSON, err := utils.CryptoJSDecrypt(embedResponse.EmbedURL, passphrase)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt embed URL: %w", err)
	}

	// Parse decrypted JSON to extract actual URL
	// The decrypted result is a JSON string containing the actual embed URL
	var embedURLData interface{}
	if err := json.Unmarshal([]byte(decryptedJSON), &embedURLData); err != nil {
		// If it's not JSON, return as-is (might be plain string)
		return strings.Trim(decryptedJSON, "\""), nil
	}

	// If it's a string, return it
	if urlStr, ok := embedURLData.(string); ok {
		return urlStr, nil
	}

	// If it's an object with "embed_url" field
	if embedMap, ok := embedURLData.(map[string]interface{}); ok {
		if url, exists := embedMap["embed_url"]; exists {
			if urlStr, ok := url.(string); ok {
				return urlStr, nil
			}
		}
	}

	return decryptedJSON, nil
}

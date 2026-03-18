package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"idlix-api/internal/models"
	"idlix-api/internal/services"
	"idlix-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// SubtitleHandler handles subtitle-related endpoints
type SubtitleHandler struct {
	service   *services.IDLIXService
	converter *utils.SubtitleConverter
	client    *utils.HTTPClient
}

// NewSubtitleHandler creates a new subtitle handler
func NewSubtitleHandler(service *services.IDLIXService, client *utils.HTTPClient) *SubtitleHandler {
	return &SubtitleHandler{
		service:   service,
		converter: utils.NewSubtitleConverter(),
		client:    client,
	}
}

// DownloadSubtitle proxies and downloads subtitle file with optional format conversion
// @Summary      Download subtitle file
// @Description  Download subtitle file with CORS support and optional format conversion
// @Tags         subtitle
// @Produce      text/plain
// @Produce      text/vtt
// @Param        url      query  string  true   "Subtitle URL"
// @Param        format   query  string  false  "Target format (srt or vtt)"
// @Param        filename query  string  false  "Custom filename"
// @Success      200  {string}  string  "Subtitle file content"
// @Failure      400  {object}  models.APIResponse  "Invalid parameters"
// @Failure      500  {object}  models.APIResponse  "Download or conversion error"
// @Router       /subtitle/download [get]
func (h *SubtitleHandler) DownloadSubtitle(c *gin.Context) {
	subtitleURL := c.Query("url")
	targetFormat := strings.ToLower(c.Query("format"))
	customFilename := c.Query("filename")

	fmt.Printf("\n📥 [SUBTITLE] Download request: URL=%s, Format=%s\n", subtitleURL, targetFormat)

	// Validate URL
	if subtitleURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"URL parameter is required",
			"INVALID_REQUEST",
			"url parameter is missing",
		))
		return
	}

	// Validate URL format
	if _, err := url.Parse(subtitleURL); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid URL format",
			"INVALID_URL",
			err.Error(),
		))
		return
	}

	// Validate target format if specified
	if targetFormat != "" && targetFormat != "srt" && targetFormat != "vtt" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid format parameter",
			"INVALID_FORMAT",
			"format must be 'srt' or 'vtt'",
		))
		return
	}

	// Download subtitle content
	fmt.Printf("📤 [SUBTITLE] Fetching subtitle from: %s\n", subtitleURL)
	
	headers := map[string]string{
		"Referer":    "https://jeniusplay.com/",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	resp, err := h.client.Get(subtitleURL, headers)
	if err != nil {
		fmt.Printf("❌ [SUBTITLE] Failed to fetch subtitle: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to download subtitle",
			"DOWNLOAD_ERROR",
			err.Error(),
		))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ [SUBTITLE] Non-200 status: %d\n", resp.StatusCode)
		c.JSON(http.StatusBadGateway, models.ErrorResponse(
			"Failed to download subtitle",
			"UPSTREAM_ERROR",
			fmt.Sprintf("upstream returned status %d", resp.StatusCode),
		))
		return
	}

	// Read subtitle content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [SUBTITLE] Failed to read response: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to read subtitle content",
			"READ_ERROR",
			err.Error(),
		))
		return
	}

	content := string(body)
	fmt.Printf("✅ [SUBTITLE] Downloaded %d bytes\n", len(content))

	// Detect current format
	detectedFormat := h.converter.DetectFormat(content)
	fmt.Printf("📝 [SUBTITLE] Detected format: %s\n", detectedFormat)

	// Convert if target format is specified and different
	finalContent := content
	finalFormat := detectedFormat

	if targetFormat != "" && targetFormat != detectedFormat {
		fmt.Printf("🔄 [SUBTITLE] Converting %s → %s\n", detectedFormat, targetFormat)
		
		converted, err := h.converter.Convert(content, detectedFormat, targetFormat)
		if err != nil {
			fmt.Printf("❌ [SUBTITLE] Conversion failed: %v\n", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse(
				"Failed to convert subtitle format",
				"CONVERSION_ERROR",
				err.Error(),
			))
			return
		}

		finalContent = converted
		finalFormat = targetFormat
		fmt.Printf("✅ [SUBTITLE] Conversion successful: %d bytes\n", len(finalContent))
	}

	// Determine filename
	filename := customFilename
	if filename == "" {
		filename = fmt.Sprintf("subtitle.%s", finalFormat)
	} else if !strings.HasSuffix(filename, "."+finalFormat) {
		filename = filename + "." + finalFormat
	}

	// Set response headers
	contentType := "text/plain"
	if finalFormat == "vtt" {
		contentType = "text/vtt"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	
	fmt.Printf("✅ [SUBTITLE] Sending %s file: %s (%d bytes)\n", finalFormat, filename, len(finalContent))

	// Send subtitle content
	c.String(http.StatusOK, finalContent)
}

// SearchSubtitles searches and filters subtitle tracks by language
// @Summary      Search subtitle tracks
// @Description  Get subtitle tracks for a video with optional language filter
// @Tags         subtitle
// @Accept       json
// @Produce      json
// @Param        url      query  string  true   "Video URL"
// @Param        language query  string  false  "Language filter (case-insensitive, partial match)"
// @Success      200  {object}  models.APIResponse{data=models.SubtitleSearchResponse}  "Subtitles found"
// @Failure      400  {object}  models.APIResponse  "Invalid parameters"
// @Failure      500  {object}  models.APIResponse  "Failed to get subtitles"
// @Router       /subtitle/search [get]
func (h *SubtitleHandler) SearchSubtitles(c *gin.Context) {
	videoURL := c.Query("url")
	languageFilter := strings.ToLower(strings.TrimSpace(c.Query("language")))

	fmt.Printf("\n🔍 [SUBTITLE] Search request: URL=%s, Language=%s\n", videoURL, languageFilter)

	// Validate URL
	if videoURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"URL parameter is required",
			"INVALID_REQUEST",
			"url parameter is missing",
		))
		return
	}

	// Get video info (includes subtitle tracks)
	videoInfo, err := h.service.GetVideoInfo(videoURL)
	if err != nil {
		fmt.Printf("❌ [SUBTITLE] Failed to get video info: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to get video information",
			"VIDEO_INFO_ERROR",
			err.Error(),
		))
		return
	}

	// Check if subtitles are available
	if videoInfo.Subtitle == nil || !videoInfo.Subtitle.Available || len(videoInfo.Subtitle.Tracks) == 0 {
		fmt.Printf("⚠️  [SUBTITLE] No subtitles available for this video\n")
		c.JSON(http.StatusOK, models.SuccessResponse(
			"No subtitles available",
			models.SubtitleSearchResponse{
				VideoID:   videoInfo.VideoID,
				VideoName: videoInfo.VideoName,
				Subtitles: []models.SubtitleTrackInfo{},
				Total:     0,
				Filtered:  false,
			},
		))
		return
	}

	// Build subtitle track info list with download URLs
	var subtitles []models.SubtitleTrackInfo
	baseURL := fmt.Sprintf("http://%s/api/v1/subtitle/download", c.Request.Host)

	for _, track := range videoInfo.Subtitle.Tracks {
		// Apply language filter if specified
		if languageFilter != "" {
			trackLangLower := strings.ToLower(track.Language)
			if !strings.Contains(trackLangLower, languageFilter) {
				continue // Skip tracks that don't match filter
			}
		}

		// Build download URL
		downloadURL := fmt.Sprintf("%s?url=%s", baseURL, url.QueryEscape(track.URL))

		subtitles = append(subtitles, models.SubtitleTrackInfo{
			Language:    track.Language,
			URL:         track.URL,
			Format:      track.Format,
			DownloadURL: downloadURL,
		})
	}

	fmt.Printf("✅ [SUBTITLE] Found %d subtitle(s) (total: %d, filtered: %v)\n", 
		len(subtitles), len(videoInfo.Subtitle.Tracks), languageFilter != "")

	// Build response
	response := models.SubtitleSearchResponse{
		VideoID:   videoInfo.VideoID,
		VideoName: videoInfo.VideoName,
		Subtitles: subtitles,
		Total:     len(subtitles),
		Filtered:  languageFilter != "",
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		fmt.Sprintf("Found %d subtitle(s)", len(subtitles)),
		response,
	))
}

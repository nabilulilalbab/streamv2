package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ProxyHandler handles proxying requests to bypass CORS
type ProxyHandler struct{}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{}
}

// ProxyM3U8 proxies M3U8 playlist and TS segment requests
// @Summary      Proxy M3U8/TS requests
// @Description  Proxy streaming requests to bypass CORS restrictions
// @Tags         proxy
// @Produce      application/x-mpegURL
// @Produce      video/MP2T
// @Param        url  query  string  true  "Target URL to proxy"
// @Success      200  "Stream content"
// @Failure      400  {object}  models.APIResponse  "Invalid URL"
// @Failure      500  {object}  models.APIResponse  "Proxy error"
// @Router       /proxy [get]
func (h *ProxyHandler) ProxyM3U8(c *gin.Context) {
	// Get target URL from query parameter
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "URL parameter is required",
		})
		return
	}

	fmt.Printf("🔄 [PROXY] Proxying request to: %s\n", targetURL)

	// Create HTTP client
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		fmt.Printf("❌ [PROXY] Failed to create request: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to create proxy request",
		})
		return
	}

	// Add headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://tv12.idlixku.com/")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ [PROXY] Request failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to fetch content",
		})
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ [PROXY] Response status: %d\n", resp.StatusCode)

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ [PROXY] Non-200 status: %d\n", resp.StatusCode)
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": fmt.Sprintf("Upstream returned status %d", resp.StatusCode),
		})
		return
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ [PROXY] Failed to read response: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to read content",
		})
		return
	}

	// Determine content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Detect content type from URL
		if strings.HasSuffix(targetURL, ".m3u8") {
			contentType = "application/x-mpegURL"
		} else if strings.HasSuffix(targetURL, ".ts") {
			contentType = "video/MP2T"
		} else {
			contentType = "application/octet-stream"
		}
	}

	// If it's M3U8 playlist, rewrite URLs to go through proxy
	if strings.Contains(contentType, "mpegURL") || strings.HasSuffix(targetURL, ".m3u8") {
		bodyStr := string(body)
		bodyStr = rewriteM3U8URLs(bodyStr, targetURL, c.Request.Host)
		body = []byte(bodyStr)
		fmt.Printf("✅ [PROXY] Rewritten M3U8 playlist (%d bytes)\n", len(body))
	}

	// Set CORS headers
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	
	// Set content type
	c.Header("Content-Type", contentType)
	
	// Cache control (optional)
	c.Header("Cache-Control", "no-cache")

	// Send response
	c.Data(http.StatusOK, contentType, body)
	fmt.Printf("✅ [PROXY] Successfully proxied %d bytes\n", len(body))
}

// rewriteM3U8URLs rewrites relative URLs in M3U8 playlist to go through proxy
func rewriteM3U8URLs(content string, originalURL string, proxyHost string) string {
	lines := strings.Split(content, "\n")
	var result []string

	// Extract base URL from original URL
	baseURL := originalURL
	if idx := strings.LastIndex(baseURL, "/"); idx != -1 {
		baseURL = baseURL[:idx+1]
	}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			result = append(result, line)
			continue
		}

		// This is a URL line
		var absoluteURL string
		
		if strings.HasPrefix(trimmedLine, "http://") || strings.HasPrefix(trimmedLine, "https://") {
			// Already absolute URL
			absoluteURL = trimmedLine
		} else if strings.HasPrefix(trimmedLine, "/") {
			// Absolute path, need to add scheme and host
			// Extract scheme and host from originalURL
			if idx := strings.Index(originalURL, "://"); idx != -1 {
				schemeAndHost := originalURL[:idx+3] // "https://"
				if idx2 := strings.Index(originalURL[idx+3:], "/"); idx2 != -1 {
					schemeAndHost += originalURL[idx+3 : idx+3+idx2]
				} else {
					schemeAndHost += originalURL[idx+3:]
				}
				absoluteURL = schemeAndHost + trimmedLine
			} else {
				absoluteURL = trimmedLine
			}
		} else {
			// Relative path
			absoluteURL = baseURL + trimmedLine
		}

		// Rewrite to proxy URL
		proxyURL := fmt.Sprintf("http://%s/api/v1/proxy?url=%s", proxyHost, absoluteURL)
		result = append(result, proxyURL)
	}

	return strings.Join(result, "\n")
}

// HandleOptions handles CORS preflight requests
func (h *ProxyHandler) HandleOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Status(http.StatusOK)
}

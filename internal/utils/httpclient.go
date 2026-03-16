package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"idlix-api/internal/models"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// HTTPClient wraps tls-client for bot detection bypass
type HTTPClient struct {
	client     tls_client.HttpClient
	userAgents []string
	baseURL    string
	timeout    time.Duration
	retry      int
}

// NewHTTPClient creates a new HTTP client with TLS fingerprinting
func NewHTTPClient(config models.IDLIXConfig) (*HTTPClient, error) {
	// Create TLS client options
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(int(config.Timeout.Seconds())),
		tls_client.WithClientProfile(profiles.Chrome_124),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithNotFollowRedirects(),
	}

	// Create client
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS client: %w", err)
	}

	// Get user agents
	userAgents := config.UserAgents
	if len(userAgents) == 0 {
		userAgents = models.GetDefaultUserAgents()
	}

	return &HTTPClient{
		client:     client,
		userAgents: userAgents,
		baseURL:    config.BaseURL,
		timeout:    config.Timeout,
		retry:      config.Retry,
	}, nil
}

// getRandomUserAgent returns a random user agent from the list
func (c *HTTPClient) getRandomUserAgent() string {
	if len(c.userAgents) == 0 {
		return models.GetDefaultUserAgents()[0]
	}
	rand.Seed(time.Now().UnixNano())
	return c.userAgents[rand.Intn(len(c.userAgents))]
}

// Get performs a GET request with retry logic
func (c *HTTPClient) Get(url string, headers map[string]string) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < c.retry; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err := c.doGet(url, headers)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		lastErr = err
		if resp != nil {
			resp.Body.Close()
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.retry, lastErr)
}

// doGet performs a single GET request
func (c *HTTPClient) doGet(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers (matching Chrome 127 exactly like Python)
	req.Header.Set("User-Agent", c.getRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	
	// CRITICAL: Client Hints headers (required for Cloudflare bypass)
	req.Header.Set("sec-ch-ua", `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	
	// Determine if this is first request or same-origin navigation
	if strings.HasPrefix(url, c.baseURL) && url != c.baseURL {
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		req.Header.Set("Referer", c.baseURL)
	} else {
		req.Header.Set("Sec-Fetch-Site", "none")
	}
	
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Cache-Control", "max-age=0")

	// Override with custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Post performs a POST request with retry logic
func (c *HTTPClient) Post(url string, headers map[string]string, body interface{}) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < c.retry; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err := c.doPost(url, headers, body)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		lastErr = err
		if resp != nil {
			resp.Body.Close()
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.retry, lastErr)
}

// doPost performs a single POST request
func (c *HTTPClient) doPost(url string, headers map[string]string, body interface{}) (*http.Response, error) {
	// Convert body to strings.Reader
	var bodyReader *strings.Reader
	switch v := body.(type) {
	case string:
		bodyReader = strings.NewReader(v)
	case []byte:
		bodyReader = strings.NewReader(string(v))
	default:
		return nil, fmt.Errorf("unsupported body type: %T", body)
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", c.getRandomUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	
	// CRITICAL: Client Hints headers (required for Cloudflare bypass)
	req.Header.Set("sec-ch-ua", `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	
	// Set origin and referer for POST requests
	if strings.HasPrefix(url, c.baseURL) {
		req.Header.Set("Origin", strings.TrimSuffix(c.baseURL, "/"))
		req.Header.Set("Referer", c.baseURL)
	}

	// Override with custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// SetCookies sets cookies for the client
func (c *HTTPClient) SetCookies(cookies []*http.Cookie) {
	// TLS client handles cookies automatically via jar
}

// GetBaseURL returns the base URL
func (c *HTTPClient) GetBaseURL() string {
	return c.baseURL
}

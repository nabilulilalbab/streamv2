package models

// VideoInfoRequest is the request body for video info endpoint
type VideoInfoRequest struct {
	URL string `json:"url" binding:"required"`
}

// VideoInfo contains complete video information
type VideoInfo struct {
	VideoID           string              `json:"video_id"`
	VideoName         string              `json:"video_name"`
	Poster            string              `json:"poster"`
	EmbedURL          string              `json:"embed_url"`
	M3U8URL           string              `json:"m3u8_url"`
	IsVariantPlaylist bool                `json:"is_variant_playlist"`
	Variants          []VariantPlaylist   `json:"variants,omitempty"`
	Subtitle          *SubtitleInfo       `json:"subtitle,omitempty"`
}

// VariantPlaylist represents a video quality variant
type VariantPlaylist struct {
	ID         string `json:"id"`
	Resolution string `json:"resolution"`
	Bandwidth  uint32 `json:"bandwidth"`
	URI        string `json:"uri"`
}

// SubtitleInfo contains subtitle information
type SubtitleInfo struct {
	Available bool   `json:"available"`
	URL       string `json:"url,omitempty"`
	Format    string `json:"format,omitempty"`
}

// DownloadRequest is the request body for download endpoint
type DownloadRequest struct {
	URL             string `json:"url" binding:"required"`
	Resolution      string `json:"resolution,omitempty"`
	IncludeSubtitle bool   `json:"include_subtitle"`
}

// DownloadJob represents a download job
type DownloadJob struct {
	JobID             string  `json:"job_id"`
	VideoName         string  `json:"video_name"`
	Resolution        string  `json:"resolution"`
	Status            string  `json:"status"` // processing, downloading, merging, completed, failed
	Progress          int     `json:"progress"`
	DownloadedSegments int    `json:"downloaded_segments,omitempty"`
	TotalSegments     int     `json:"total_segments,omitempty"`
	FileSize          string  `json:"file_size,omitempty"`
	FilePath          string  `json:"file_path,omitempty"`
	ErrorMessage      string  `json:"error_message,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at,omitempty"`
}

// EncryptedEmbed represents encrypted embed response from IDLIX
type EncryptedEmbed struct {
	EmbedURL string `json:"embed_url"` // This is a JSON string containing encrypted data
	Type     string `json:"type"`
	Key      string `json:"key"`
}

// EncryptedData represents the inner encrypted data structure
type EncryptedData struct {
	CT string `json:"ct"` // Ciphertext
	IV string `json:"iv"` // Initialization vector
	S  string `json:"s"`  // Salt
	M  string `json:"m"`  // M parameter for passphrase generation
}

// JeniusPlayVideoResponse represents response from JeniusPlay API
type JeniusPlayVideoResponse struct {
	VideoSource string `json:"videoSource"`
	SecSrc      string `json:"secSrc,omitempty"`
	Tracks      string `json:"tracks,omitempty"`
}

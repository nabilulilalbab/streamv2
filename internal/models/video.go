package models

// VideoInfoRequest is the request body for video info endpoint
type VideoInfoRequest struct {
	URL string `json:"url" binding:"required" example:"https://tv12.idlixku.com/movie/crime-101-2026/"`
}

// VideoInfo contains complete video information
type VideoInfo struct {
	VideoID           string              `json:"video_id" example:"163426"`
	VideoName         string              `json:"video_name" example:"Crime 101 (2026)"`
	Poster            string              `json:"poster" example:"https://image.tmdb.org/t/p/w185/poster.jpg"`
	EmbedURL          string              `json:"embed_url" example:"https://jeniusplay.com/video/hash123"`
	M3U8URL           string              `json:"m3u8_url" example:"https://jeniusplay.com/cdn/hls/hash123/master.m3u8"`
	IsVariantPlaylist bool                `json:"is_variant_playlist" example:"true"`
	Variants          []VariantPlaylist   `json:"variants,omitempty"`
	Subtitle          *SubtitleInfo       `json:"subtitle,omitempty"`
}

// VariantPlaylist represents a video quality variant
type VariantPlaylist struct {
	ID         string `json:"id" example:"0"`
	Resolution string `json:"resolution" example:"1920x1080"`
	Bandwidth  uint32 `json:"bandwidth" example:"1510000"`
	URI        string `json:"uri" example:"https://jeniusplay.com/hls/hash123/1080p.m3u8"`
}

// SubtitleInfo contains subtitle information
type SubtitleInfo struct {
	Available bool            `json:"available" example:"true"`
	Tracks    []SubtitleTrack `json:"tracks,omitempty"`
}

// SubtitleTrack represents a single subtitle track with language
type SubtitleTrack struct {
	Language string `json:"language" example:"Bahasa"`
	URL      string `json:"url" example:"https://g5.wiseacademia.asia/r/xyz.jpg"`
	Format   string `json:"format" example:"srt"`
}

// SubtitleSearchResponse contains search results for subtitles
type SubtitleSearchResponse struct {
	VideoID   string              `json:"video_id" example:"163426"`
	VideoName string              `json:"video_name" example:"Crime 101 (2026)"`
	Subtitles []SubtitleTrackInfo `json:"subtitles"`
	Total     int                 `json:"total" example:"2"`
	Filtered  bool                `json:"filtered" example:"true"`
}

// SubtitleTrackInfo contains subtitle track info with download URL
type SubtitleTrackInfo struct {
	Language    string `json:"language" example:"Bahasa"`
	URL         string `json:"url" example:"https://g5.wiseacademia.asia/r/xyz.jpg"`
	Format      string `json:"format" example:"srt"`
	DownloadURL string `json:"download_url" example:"/api/v1/subtitle/download?url=..."`
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
	VideoSource  string `json:"videoSource"`
	SecuredLink  string `json:"securedLink,omitempty"`  // Secured M3U8 URL with auth
	SecSrc       string `json:"secSrc,omitempty"`
	Tracks       string `json:"tracks,omitempty"`
}

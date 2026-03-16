package models

import "time"

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	IDLIX    IDLIXConfig    `yaml:"idlix"`
	Jenius   JeniusConfig   `yaml:"jenius"`
	Download DownloadConfig `yaml:"download"`
	FFmpeg   FFmpegConfig   `yaml:"ffmpeg"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `yaml:"port" env:"PORT" envDefault:"8080"`
	Host string `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Mode string `yaml:"mode" env:"GIN_MODE" envDefault:"release"`
}

// IDLIXConfig holds IDLIX scraper configuration
type IDLIXConfig struct {
	BaseURL    string        `yaml:"base_url" envDefault:"https://tv12.idlixku.com/"`
	Timeout    time.Duration `yaml:"timeout" envDefault:"30s"`
	Retry      int           `yaml:"retry" envDefault:"3"`
	UserAgents []string      `yaml:"user_agents"`
}

// JeniusConfig holds JeniusPlay configuration
type JeniusConfig struct {
	BaseURL string        `yaml:"base_url" envDefault:"https://jeniusplay.com/"`
	Timeout time.Duration `yaml:"timeout" envDefault:"30s"`
}

// DownloadConfig holds download configuration
type DownloadConfig struct {
	MaxConcurrent  int           `yaml:"max_concurrent" envDefault:"10"`
	SegmentWorkers int           `yaml:"segment_workers" envDefault:"50"`
	OutputDir      string        `yaml:"output_dir" envDefault:"./downloads"`
	TempDir        string        `yaml:"temp_dir" envDefault:"./tmp"`
	CleanupAfter   time.Duration `yaml:"cleanup_after" envDefault:"24h"`
}

// FFmpegConfig holds FFmpeg configuration
type FFmpegConfig struct {
	Path    string   `yaml:"path" envDefault:"ffmpeg"`
	Options []string `yaml:"options"`
}

// GetDefaultUserAgents returns default user agents
func GetDefaultUserAgents() []string {
	return []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	}
}

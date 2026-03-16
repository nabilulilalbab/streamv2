package services

import (
	"fmt"

	"idlix-api/internal/models"
	"idlix-api/internal/repositories"
	"idlix-api/internal/utils"
)

// IDLIXService handles business logic for IDLIX operations
type IDLIXService struct {
	idlixRepo   *repositories.IDLIXRepository
	jeniusRepo  *repositories.JeniusRepository
	m3u8Parser  *utils.M3U8Parser
}

// NewIDLIXService creates a new IDLIX service
func NewIDLIXService(
	idlixRepo *repositories.IDLIXRepository,
	jeniusRepo *repositories.JeniusRepository,
	m3u8Parser *utils.M3U8Parser,
) *IDLIXService {
	return &IDLIXService{
		idlixRepo:  idlixRepo,
		jeniusRepo: jeniusRepo,
		m3u8Parser: m3u8Parser,
	}
}

// GetFeaturedMovies retrieves featured movies from IDLIX
func (s *IDLIXService) GetFeaturedMovies() (*models.FeaturedMoviesResponse, error) {
	movies, err := s.idlixRepo.GetFeaturedMovies()
	if err != nil {
		return nil, fmt.Errorf("failed to get featured movies: %w", err)
	}

	return &models.FeaturedMoviesResponse{
		Movies: movies,
	}, nil
}

// GetVideoInfo retrieves complete video information
func (s *IDLIXService) GetVideoInfo(movieURL string) (*models.VideoInfo, error) {
	// Step 1: Get video data
	videoID, videoName, poster, err := s.idlixRepo.GetVideoData(movieURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get video data: %w", err)
	}

	// Step 2: Get and decrypt embed URL
	embedURL, err := s.idlixRepo.GetEmbedURL(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed URL: %w", err)
	}

	// Step 3: Extract embed hash from URL
	embedHash, err := s.jeniusRepo.ExtractEmbedHash(embedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract embed hash: %w", err)
	}

	// Step 4: Get video source from JeniusPlay
	videoSource, err := s.jeniusRepo.GetVideoSource(embedHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get video source: %w", err)
	}

	// Step 5: Convert MP4 URL to M3U8
	m3u8URL := utils.ConvertMP4ToM3U8(videoSource.VideoSource)

	// Step 6: Parse M3U8 playlist and get variants
	variants, isVariant, err := s.m3u8Parser.ParseMasterPlaylist(m3u8URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse M3U8 playlist: %w", err)
	}

	// Step 7: Get subtitle URL (optional, don't fail if not found)
	var subtitle *models.SubtitleInfo
	subtitleURL, err := s.jeniusRepo.GetSubtitleURL(embedHash)
	if err == nil && subtitleURL != "" {
		subtitle = &models.SubtitleInfo{
			Available: true,
			URL:       subtitleURL,
			Format:    "vtt",
		}
	} else {
		subtitle = &models.SubtitleInfo{
			Available: false,
		}
	}

	// Return complete video info
	return &models.VideoInfo{
		VideoID:           videoID,
		VideoName:         videoName,
		Poster:            poster,
		EmbedURL:          embedURL,
		M3U8URL:           m3u8URL,
		IsVariantPlaylist: isVariant,
		Variants:          variants,
		Subtitle:          subtitle,
	}, nil
}

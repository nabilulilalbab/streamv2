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
	fmt.Printf("\n🎬 [SERVICE] Starting GetVideoInfo for URL: %s\n", movieURL)
	
	// Step 1: Get video data
	fmt.Println("📍 [SERVICE] Step 1: Getting video data...")
	videoID, videoName, poster, err := s.idlixRepo.GetVideoData(movieURL)
	if err != nil {
		fmt.Printf("❌ [SERVICE] Step 1 FAILED: %v\n", err)
		return nil, fmt.Errorf("failed to get video data: %w", err)
	}
	fmt.Printf("✅ [SERVICE] Step 1 SUCCESS: VideoID=%s, VideoName=%s\n", videoID, videoName)

	// Step 2: Get and decrypt embed URL
	fmt.Println("📍 [SERVICE] Step 2: Getting and decrypting embed URL...")
	embedURL, err := s.idlixRepo.GetEmbedURL(videoID)
	if err != nil {
		fmt.Printf("❌ [SERVICE] Step 2 FAILED: %v\n", err)
		return nil, fmt.Errorf("failed to get embed URL: %w", err)
	}
	fmt.Printf("✅ [SERVICE] Step 2 SUCCESS: EmbedURL=%s\n", embedURL)

	// Step 3: Extract embed hash from URL
	fmt.Println("📍 [SERVICE] Step 3: Extracting embed hash from URL...")
	embedHash, err := s.jeniusRepo.ExtractEmbedHash(embedURL)
	if err != nil {
		fmt.Printf("❌ [SERVICE] Step 3 FAILED: %v\n", err)
		return nil, fmt.Errorf("failed to extract embed hash: %w", err)
	}
	fmt.Printf("✅ [SERVICE] Step 3 SUCCESS: EmbedHash=%s\n", embedHash)

	// Step 4: Get video source from JeniusPlay
	fmt.Println("📍 [SERVICE] Step 4: Getting video source from JeniusPlay...")
	videoSource, err := s.jeniusRepo.GetVideoSource(embedHash)
	if err != nil {
		fmt.Printf("❌ [SERVICE] Step 4 FAILED: %v\n", err)
		return nil, fmt.Errorf("failed to get video source: %w", err)
	}
	fmt.Printf("✅ [SERVICE] Step 4 SUCCESS: VideoSource=%s\n", videoSource.VideoSource)

	// Step 5: Convert MP4 URL to M3U8
	fmt.Println("📍 [SERVICE] Step 5: Converting MP4 URL to M3U8...")
	m3u8URL := utils.ConvertMP4ToM3U8(videoSource.VideoSource)
	fmt.Printf("✅ [SERVICE] Step 5 SUCCESS: M3U8URL=%s\n", m3u8URL)

	// Step 6: Parse M3U8 playlist and get variants
	fmt.Println("📍 [SERVICE] Step 6: Parsing M3U8 playlist...")
	variants, isVariant, err := s.m3u8Parser.ParseMasterPlaylist(m3u8URL)
	if err != nil {
		fmt.Printf("❌ [SERVICE] Step 6 FAILED: %v\n", err)
		return nil, fmt.Errorf("failed to parse M3U8 playlist: %w", err)
	}
	fmt.Printf("✅ [SERVICE] Step 6 SUCCESS: IsVariant=%v, Variants=%d\n", isVariant, len(variants))

	// Step 7: Get subtitle URL (optional, don't fail if not found)
	fmt.Println("📍 [SERVICE] Step 7: Getting subtitle URL (optional)...")
	var subtitle *models.SubtitleInfo
	subtitleURL, err := s.jeniusRepo.GetSubtitleURL(embedHash)
	if err == nil && subtitleURL != "" {
		subtitle = &models.SubtitleInfo{
			Available: true,
			URL:       subtitleURL,
			Format:    "vtt",
		}
		fmt.Printf("✅ [SERVICE] Step 7 SUCCESS: SubtitleURL=%s\n", subtitleURL)
	} else {
		subtitle = &models.SubtitleInfo{
			Available: false,
		}
		fmt.Printf("⚠️  [SERVICE] Step 7: No subtitle available (error: %v)\n", err)
	}

	fmt.Println("🎉 [SERVICE] GetVideoInfo completed successfully!")
	
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

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"idlix-api/internal/models"
	"idlix-api/internal/repositories"
	"idlix-api/internal/utils"
)

func testScraper() {
	fmt.Println("🧪 Testing IDLIX Scraper...")
	fmt.Println("=" + strings.Repeat("=", 50))

	// Create config
	config := models.IDLIXConfig{
		BaseURL: "https://tv12.idlixku.com/",
		Timeout: 30 * time.Second,
		Retry:   3,
	}

	// Create HTTP client
	client, err := utils.NewHTTPClient(config)
	if err != nil {
		log.Fatalf("❌ Failed to create HTTP client: %v", err)
	}
	fmt.Println("✅ HTTP Client created successfully")

	// Create repository
	repo := repositories.NewIDLIXRepository(client)
	fmt.Println("✅ IDLIX Repository created")

	// Test GetFeaturedMovies
	fmt.Println("\n📡 Fetching featured movies...")
	movies, err := repo.GetFeaturedMovies()
	if err != nil {
		log.Fatalf("❌ Failed to get featured movies: %v", err)
	}

	fmt.Printf("✅ Found %d featured movies\n\n", len(movies))

	// Display first 5 movies
	displayCount := 5
	if len(movies) < displayCount {
		displayCount = len(movies)
	}

	for i := 0; i < displayCount; i++ {
		movie := movies[i]
		fmt.Printf("Movie %d:\n", i+1)
		fmt.Printf("  Title:  %s\n", movie.Title)
		fmt.Printf("  Year:   %s\n", movie.Year)
		fmt.Printf("  Type:   %s\n", movie.Type)
		fmt.Printf("  URL:    %s\n", movie.URL)
		fmt.Printf("  Poster: %s\n", movie.Poster)
		fmt.Println()
	}

	// Test GetVideoData with first movie
	if len(movies) > 0 {
		fmt.Println("📡 Testing GetVideoData with first movie...")
		firstMovie := movies[0]
		
		videoID, videoName, poster, err := repo.GetVideoData(firstMovie.URL)
		if err != nil {
			log.Fatalf("❌ Failed to get video data: %v", err)
		}

		fmt.Println("✅ Video data retrieved successfully")
		fmt.Printf("  Video ID:   %s\n", videoID)
		fmt.Printf("  Video Name: %s\n", videoName)
		fmt.Printf("  Poster:     %s\n", poster)
	}

	fmt.Println("\n" + strings.Repeat("=", 52))
	fmt.Println("✅ All scraper tests passed!")
}

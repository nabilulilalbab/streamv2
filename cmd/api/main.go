package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"idlix-api/internal/handlers"
	"idlix-api/internal/models"
	"idlix-api/internal/repositories"
	"idlix-api/internal/services"
	"idlix-api/internal/utils"
	"idlix-api/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// For now, run the scraper test
	if len(os.Args) > 1 && os.Args[1] == "test-scraper" {
		testScraper()
		return
	}

	// Start API server
	startServer()
}

func startServer() {
	fmt.Println("🚀 Starting IDLIX API Server...")

	// Load configuration
	config := models.Config{
		Server: models.ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		IDLIX: models.IDLIXConfig{
			BaseURL: getEnv("IDLIX_BASE_URL", "https://tv12.idlixku.com/"),
			Timeout: 30 * time.Second,
			Retry:   3,
		},
	}

	// Set Gin mode
	gin.SetMode(config.Server.Mode)

	// Create HTTP client
	httpClient, err := utils.NewHTTPClient(config.IDLIX)
	if err != nil {
		log.Fatalf("❌ Failed to create HTTP client: %v", err)
	}
	fmt.Println("✅ HTTP Client initialized")

	// Create repositories
	idlixRepo := repositories.NewIDLIXRepository(httpClient)
	fmt.Println("✅ IDLIX Repository initialized")

	jeniusRepo := repositories.NewJeniusRepository(httpClient, "https://jeniusplay.com/")
	fmt.Println("✅ JeniusPlay Repository initialized")

	// Create utilities
	m3u8Parser := utils.NewM3U8Parser(httpClient)
	fmt.Println("✅ M3U8 Parser initialized")

	// Create services
	idlixService := services.NewIDLIXService(idlixRepo, jeniusRepo, m3u8Parser)
	fmt.Println("✅ IDLIX Service initialized")

	// Create handlers
	featuredHandler := handlers.NewFeaturedHandler(idlixService)
	videoHandler := handlers.NewVideoHandler(idlixService)
	fmt.Println("✅ Handlers initialized")

	// Setup router
	router := gin.New()
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Featured movies
		v1.GET("/featured", featuredHandler.GetFeatured)

		// Video info
		v1.POST("/video/info", videoHandler.GetVideoInfo)

		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"version": "1.0.0",
				"message": "IDLIX API is running",
			})
		})
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
	fmt.Printf("\n✅ Server is running on http://%s\n", addr)
	fmt.Println("📡 API Endpoints:")
	fmt.Printf("   GET  http://localhost:%s/api/v1/health\n", config.Server.Port)
	fmt.Printf("   GET  http://localhost:%s/api/v1/featured\n", config.Server.Port)
	fmt.Printf("   POST http://localhost:%s/api/v1/video/info\n", config.Server.Port)
	fmt.Println()

	if err := router.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

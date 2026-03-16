package handlers

import (
	"net/http"

	"idlix-api/internal/models"
	"idlix-api/internal/services"

	"github.com/gin-gonic/gin"
)

// VideoHandler handles video info endpoint
type VideoHandler struct {
	service *services.IDLIXService
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(service *services.IDLIXService) *VideoHandler {
	return &VideoHandler{
		service: service,
	}
}

// GetVideoInfo godoc
// @Summary      Get video information
// @Description  Get complete video information including M3U8 streams, quality variants, and subtitles
// @Tags         videos
// @Accept       json
// @Produce      json
// @Param        request  body      models.VideoInfoRequest  true  "Video URL request"
// @Success      200      {object}  models.APIResponse{data=models.VideoInfo}  "Video info retrieved successfully"
// @Failure      400      {object}  models.APIResponse  "Invalid request body"
// @Failure      500      {object}  models.APIResponse  "Internal server error"
// @Router       /video/info [post]
func (h *VideoHandler) GetVideoInfo(c *gin.Context) {
	var req models.VideoInfoRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request body",
			"INVALID_REQUEST",
			err.Error(),
		))
		return
	}

	// Get video info
	videoInfo, err := h.service.GetVideoInfo(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to get video info",
			"VIDEO_INFO_ERROR",
			err.Error(),
		))
		return
	}

	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		"Video info retrieved successfully",
		videoInfo,
	))
}

package handlers

import (
	"net/http"

	"idlix-api/internal/models"
	"idlix-api/internal/services"

	"github.com/gin-gonic/gin"
)

// FeaturedHandler handles featured movies endpoint
type FeaturedHandler struct {
	service *services.IDLIXService
}

// NewFeaturedHandler creates a new featured handler
func NewFeaturedHandler(service *services.IDLIXService) *FeaturedHandler {
	return &FeaturedHandler{
		service: service,
	}
}

// GetFeatured godoc
// @Summary      Get featured movies
// @Description  Retrieve list of featured movies from IDLIX homepage (excluding TV series)
// @Tags         movies
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.APIResponse{data=models.FeaturedMoviesResponse}  "Featured movies retrieved successfully"
// @Failure      500  {object}  models.APIResponse  "Internal server error"
// @Router       /featured [get]
func (h *FeaturedHandler) GetFeatured(c *gin.Context) {
	// Get featured movies
	response, err := h.service.GetFeaturedMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to get featured movies",
			"FEATURED_FETCH_ERROR",
			err.Error(),
		))
		return
	}

	// Return success response
	c.JSON(http.StatusOK, models.SuccessResponse(
		"Featured movies retrieved successfully",
		response,
	))
}

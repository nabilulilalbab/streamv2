package models

// Movie represents a movie from IDLIX featured list
type Movie struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Year   string `json:"year"`
	Type   string `json:"type"`
	Poster string `json:"poster"`
}

// FeaturedMoviesResponse is the response for featured movies endpoint
type FeaturedMoviesResponse struct {
	Movies []Movie `json:"movies"`
}

package models

// Movie represents a movie from IDLIX featured list
type Movie struct {
	URL    string `json:"url" example:"https://tv12.idlixku.com/movie/example-2024/"`
	Title  string `json:"title" example:"Example Movie (2024)"`
	Year   string `json:"year" example:"2024"`
	Type   string `json:"type" example:"movie"`
	Poster string `json:"poster" example:"https://image.tmdb.org/t/p/w185/poster.jpg"`
}

// FeaturedMoviesResponse is the response for featured movies endpoint
type FeaturedMoviesResponse struct {
	Movies []Movie `json:"movies"`
}

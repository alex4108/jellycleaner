package radarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client handles communication with the Radarr API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Movie represents a movie in Radarr
type Movie struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	TMDBID   int    `json:"tmdbId"`
	FilePath string `json:"path"`
}

// NewClient creates a new Radarr client
func NewClient(baseURL, apiKey string) (*Client, error) {
	// Ensure baseURL doesn't end with a slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GetMovieByTMDBID gets a movie by its TMDB ID
func (c *Client) GetMovieByTMDBID(tmdbID string) (*Movie, error) {
	// Convert string to int
	tmdbIDInt, err := strconv.Atoi(tmdbID)
	if err != nil {
		return nil, fmt.Errorf("invalid TMDB ID format: %v", err)
	}

	// Get all movies
	movies, err := c.GetAllMovies()
	if err != nil {
		return nil, err
	}

	// Find the movie with matching TMDB ID
	for _, movie := range movies {
		if movie.TMDBID == tmdbIDInt {
			return &movie, nil
		}
	}

	return nil, fmt.Errorf("movie with TMDB ID %s not found", tmdbID)
}

// GetAllMovies gets all movies from Radarr
func (c *Client) GetAllMovies() ([]Movie, error) {
	endpoint := "/api/v3/movie"
	var movies []Movie

	if err := c.get(endpoint, &movies); err != nil {
		return nil, err
	}

	return movies, nil
}

// DeleteMovie deletes a movie from Radarr
func (c *Client) DeleteMovie(tmdbID string) error {
	// First get the Radarr movie ID from TMDB ID
	movie, err := c.GetMovieByTMDBID(tmdbID)
	if err != nil {
		return err
	}

	// Delete the movie
	endpoint := fmt.Sprintf("/api/v3/movie/%d", movie.ID)
	
	// Add query parameters for deletion options
	queryParams := url.Values{}
	queryParams.Add("deleteFiles", "true")  // Delete the movie files
	queryParams.Add("addImportExclusion", "false") // Don't add to import exclusions
	
	endpoint = endpoint + "?" + queryParams.Encode()
	
	return c.delete(endpoint, nil)
}

// HTTP helpers
func (c *Client) get(endpoint string, response interface{}) error {
	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, response)
}

func (c *Client) post(endpoint string, body interface{}, response interface{}) error {
	var bodyJSON []byte
	var err error
	
	if body != nil {
		bodyJSON, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, strings.NewReader(string(bodyJSON)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, response)
}

func (c *Client) delete(endpoint string, response interface{}) error {
	req, err := http.NewRequest("DELETE", c.baseURL+endpoint, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, response)
}

func (c *Client) doRequest(req *http.Request, response interface{}) error {
	// Add API key to all requests
	q := req.URL.Query()
	q.Add("apikey", c.apiKey)
	req.URL.RawQuery = q.Encode()
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	if response != nil {
		return json.NewDecoder(resp.Body).Decode(response)
	}

	return nil
}
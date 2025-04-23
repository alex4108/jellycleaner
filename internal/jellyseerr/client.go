package jellyseerr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client handles communication with the Jellyseerr API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// MediaRequest represents a media request in Jellyseerr
type MediaRequest struct {
	ID         int    `json:"id"`
	MediaType  string `json:"media"`  // "movie" or "tv"
	MediaID    int    `json:"mediaId"` // TMDB ID for movies or TVDB ID for series
	Status     int    `json:"status"`
	RequestedBy struct {
		ID   int    `json:"id"`
		Name string `json:"displayName"`
	} `json:"requestedBy"`
}

// NewClient creates a new Jellyseerr client
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

// GetAllRequests gets all media requests from Jellyseerr
func (c *Client) GetAllRequests() ([]MediaRequest, error) {
	endpoint := "/api/v1/request?take=100&skip=0&filter=all&sort=added"
	var response struct {
		Results  []MediaRequest `json:"results"`
		PageInfo struct {
			PageSize int `json:"pageSize"`
			Results  int `json:"results"`
			Pages    int `json:"pages"`
			Page     int `json:"page"`
		} `json:"pageInfo"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	// If there are multiple pages, fetch them all
	allRequests := response.Results
	
	for page := 2; page <= response.PageInfo.Pages; page++ {
		endpoint := fmt.Sprintf("/api/v1/request?take=100&skip=%d&filter=all&sort=added", (page-1)*100)
		
		if err := c.get(endpoint, &response); err != nil {
			return nil, err
		}
		
		allRequests = append(allRequests, response.Results...)
	}

	return allRequests, nil
}

// DeleteMovieRequest deletes a movie request by TMDB ID
func (c *Client) DeleteMovieRequest(tmdbID string) error {
	// Convert string to int
	tmdbIDInt, err := strconv.Atoi(tmdbID)
	if err != nil {
		return fmt.Errorf("invalid TMDB ID format: %v", err)
	}

	// Get all requests
	requests, err := c.GetAllRequests()
	if err != nil {
		return err
	}

	// Find the request with matching TMDB ID
	for _, request := range requests {
		if request.MediaType == "movie" && request.MediaID == tmdbIDInt {
			// Delete the request
			return c.DeleteRequest(request.ID)
		}
	}

	return fmt.Errorf("movie request with TMDB ID %s not found", tmdbID)
}

// DeleteSeriesRequest deletes a series request by TVDB ID
func (c *Client) DeleteSeriesRequest(tvdbID string) error {
	// Convert string to int
	tvdbIDInt, err := strconv.Atoi(tvdbID)
	if err != nil {
		return fmt.Errorf("invalid TVDB ID format: %v", err)
	}

	// Get all requests
	requests, err := c.GetAllRequests()
	if err != nil {
		return err
	}

	// Find the request with matching TVDB ID
	for _, request := range requests {
		if request.MediaType == "tv" && request.MediaID == tvdbIDInt {
			// Delete the request
			return c.DeleteRequest(request.ID)
		}
	}

	return fmt.Errorf("series request with TVDB ID %s not found", tvdbID)
}

// DeleteRequest deletes a request by its ID
func (c *Client) DeleteRequest(requestID int) error {
	endpoint := fmt.Sprintf("/api/v1/request/%d", requestID)
	return c.delete(endpoint, nil)
}

// DeleteMediaFromJellyseerr removes media from Jellyseerr when it's deleted from Sonarr/Radarr
func (c *Client) DeleteMediaFromJellyseerr(mediaType string, externalID string) error {
	if mediaType == "movie" {
		return c.DeleteMovieRequest(externalID)
	} else if mediaType == "tv" || mediaType == "series" {
		return c.DeleteSeriesRequest(externalID)
	}
	
	return fmt.Errorf("unsupported media type: %s", mediaType)
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
	req.Header.Set("X-Api-Key", c.apiKey)
	
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
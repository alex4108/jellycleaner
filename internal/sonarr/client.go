package sonarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client handles communication with the Sonarr API
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Series represents a TV series in Sonarr
type Series struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	TVDBID  int    `json:"tvdbId"`
	Path    string `json:"path"`
}

// NewClient creates a new Sonarr client
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

// GetSeriesByTVDBID gets a series by its TVDB ID
func (c *Client) GetSeriesByTVDBID(tvdbID string) (*Series, error) {
	// Convert string to int
	tvdbIDInt, err := strconv.Atoi(tvdbID)
	if err != nil {
		return nil, fmt.Errorf("invalid TVDB ID format: %v", err)
	}

	// Get all series
	allSeries, err := c.GetAllSeries()
	if err != nil {
		return nil, err
	}

	// Find the series with matching TVDB ID
	for _, series := range allSeries {
		if series.TVDBID == tvdbIDInt {
			return &series, nil
		}
	}

	return nil, fmt.Errorf("series with TVDB ID %s not found", tvdbID)
}

// GetAllSeries gets all series from Sonarr
func (c *Client) GetAllSeries() ([]Series, error) {
	endpoint := "/api/v3/series"
	var series []Series

	if err := c.get(endpoint, &series); err != nil {
		return nil, err
	}

	return series, nil
}

// DeleteSeries deletes a series from Sonarr
func (c *Client) DeleteSeries(tvdbID string) error {
	// First get the Sonarr series ID from TVDB ID
	series, err := c.GetSeriesByTVDBID(tvdbID)
	if err != nil {
		return err
	}

	// Delete the series
	endpoint := fmt.Sprintf("/api/v3/series/%d", series.ID)
	
	// Add query parameters for deletion options
	queryParams := url.Values{}
	queryParams.Add("deleteFiles", "true")  // Delete the series files
	queryParams.Add("addImportListExclusion", "false") // Don't add to import list exclusions
	
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
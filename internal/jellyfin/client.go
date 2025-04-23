package jellyfin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client handles communication with the Jellyfin API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Item represents a media item in Jellyfin
type Item struct {
	ID         string
	Name       string
	Type       string // "Series" or "Movie"
	ExternalID string // TVDB/TMDB ID
}

// NewClient creates a new Jellyfin client
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

// GetLibraryItems returns all items in a specific library
func (c *Client) GetLibraryItems(libraryName string) ([]Item, error) {
	// First, get the library ID by name
	libraryID, err := c.getLibraryIDByName(libraryName)
	if err != nil {
		return nil, err
	}

	// Now get all items in the library
	endpoint := fmt.Sprintf("/Items?ParentId=%s", url.QueryEscape(libraryID))
	var response struct {
		Items []struct {
			ID            string   `json:"Id"`
			Name          string   `json:"Name"`
			Type          string   `json:"Type"`
			ProviderIDs   map[string]string `json:"ProviderIds"`
		} `json:"Items"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	var items []Item
	for _, item := range response.Items {
		externalID := ""
		if item.Type == "Series" && item.ProviderIDs["Tvdb"] != "" {
			externalID = item.ProviderIDs["Tvdb"]
		} else if item.Type == "Movie" && item.ProviderIDs["Tmdb"] != "" {
			externalID = item.ProviderIDs["Tmdb"]
		}

		items = append(items, Item{
			ID:         item.ID,
			Name:       item.Name,
			Type:       item.Type,
			ExternalID: externalID,
		})
	}

	return items, nil
}

// IsWatchedByAllUsers checks if the item has been watched by all users
func (c *Client) IsWatchedByAllUsers(itemID string) (bool, error) {
	// First, get all users
	users, err := c.getUsers()
	if err != nil {
		return false, err
	}

	// Check if the item is played for each user
	for _, user := range users {
		played, err := c.isItemPlayedByUser(itemID, user.ID)
		if err != nil {
			return false, err
		}
		if !played {
			return false, nil // Not watched by at least one user
		}
	}

	return true, nil // Watched by all users
}

// GetItemAddedDate returns the date when the item was added to Jellyfin
func (c *Client) GetItemAddedDate(itemID string) (time.Time, error) {
	endpoint := fmt.Sprintf("/Items/%s", url.QueryEscape(itemID))
	var response struct {
		DateCreated string `json:"DateCreated"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, response.DateCreated)
}

// IsInPlaylist checks if an item is in a specific playlist
func (c *Client) IsInPlaylist(itemID, playlistName string) bool {
	playlistID, err := c.getPlaylistIDByName(playlistName)
	if err != nil {
		return false
	}

	endpoint := fmt.Sprintf("/Playlists/%s/Items", url.QueryEscape(playlistID))
	var response struct {
		Items []struct {
			ID string `json:"Id"`
		} `json:"Items"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return false
	}

	for _, item := range response.Items {
		if item.ID == itemID {
			return true
		}
	}

	return false
}

// AddToPlaylist adds an item to a playlist
func (c *Client) AddToPlaylist(itemID, playlistName string) error {
	// Get or create the playlist
	playlistID, err := c.getOrCreatePlaylist(playlistName)
	if err != nil {
		return err
	}

	// Add item to playlist
	endpoint := fmt.Sprintf("/Playlists/%s/Items?Ids=%s", url.QueryEscape(playlistID), url.QueryEscape(itemID))
	return c.post(endpoint, nil, nil)
}

// RemoveFromPlaylist removes an item from a playlist
func (c *Client) RemoveFromPlaylist(itemID, playlistName string) error {
	playlistID, err := c.getPlaylistIDByName(playlistName)
	if err != nil {
		return err
	}

	// Get item index in playlist
	itemIndex, err := c.getItemIndexInPlaylist(itemID, playlistID)
	if err != nil {
		return err
	}

	// Remove item from playlist
	endpoint := fmt.Sprintf("/Playlists/%s/Items?EntryIds=%d", url.QueryEscape(playlistID), itemIndex)
	return c.delete(endpoint, nil)
}

// GetPlaylistItems gets all items in a specific playlist
func (c *Client) GetPlaylistItems(playlistName string) ([]Item, error) {
	playlistID, err := c.getPlaylistIDByName(playlistName)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/Playlists/%s/Items", url.QueryEscape(playlistID))
	var response struct {
		Items []struct {
			ID            string   `json:"Id"`
			Name          string   `json:"Name"`
			Type          string   `json:"Type"`
			ProviderIDs   map[string]string `json:"ProviderIds"`
		} `json:"Items"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	var items []Item
	for _, item := range response.Items {
		externalID := ""
		if item.Type == "Series" && item.ProviderIDs["Tvdb"] != "" {
			externalID = item.ProviderIDs["Tvdb"]
		} else if item.Type == "Movie" && item.ProviderIDs["Tmdb"] != "" {
			externalID = item.ProviderIDs["Tmdb"]
		}

		items = append(items, Item{
			ID:         item.ID,
			Name:       item.Name,
			Type:       item.Type,
			ExternalID: externalID,
		})
	}

	return items, nil
}

// AddTag adds a tag to an item
func (c *Client) AddTag(itemID, tag string) error {
	// Get current tags
	currentTags, err := c.getTags(itemID)
	if err != nil {
		return err
	}

	// Add new tag if not present
	tags := append(currentTags, tag)
	return c.updateTags(itemID, tags)
}

// RemoveTag removes a tag from an item
func (c *Client) RemoveTag(itemID, tag string) error {
	// Get current tags
	currentTags, err := c.getTags(itemID)
	if err != nil {
		return err
	}

	// Remove tag
	var newTags []string
	for _, t := range currentTags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	return c.updateTags(itemID, newTags)
}

// GetExpirationTags returns all expiration tags for an item
func (c *Client) GetExpirationTags(itemID string) []string {
	tags, err := c.getTags(itemID)
	if err != nil {
		return nil
	}

	var expirationTags []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, "Expires: ") {
			expirationTags = append(expirationTags, tag)
		}
	}

	return expirationTags
}

// Helper methods
func (c *Client) getLibraryIDByName(name string) (string, error) {
	endpoint := "/Library/MediaFolders"
	var response struct {
		Items []struct {
			ID   string `json:"Id"`
			Name string `json:"Name"`
		} `json:"Items"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return "", err
	}

	for _, item := range response.Items {
		if item.Name == name {
			return item.ID, nil
		}
	}

	return "", fmt.Errorf("library not found: %s", name)
}

func (c *Client) getUsers() ([]struct{ ID string }, error) {
	endpoint := "/Users"
	var users []struct {
		ID string `json:"Id"`
	}

	if err := c.get(endpoint, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *Client) isItemPlayedByUser(itemID, userID string) (bool, error) {
	endpoint := fmt.Sprintf("/Users/%s/Items/%s", url.QueryEscape(userID), url.QueryEscape(itemID))
	var response struct {
		UserData struct {
			Played bool `json:"Played"`
		} `json:"UserData"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return false, err
	}

	return response.UserData.Played, nil
}

func (c *Client) getPlaylistIDByName(name string) (string, error) {
	endpoint := "/Playlists"
	var response struct {
		Items []struct {
			ID   string `json:"Id"`
			Name string `json:"Name"`
		} `json:"Items"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return "", err
	}

	for _, item := range response.Items {
		if item.Name == name {
			return item.ID, nil
		}
	}

	return "", fmt.Errorf("playlist not found: %s", name)
}

func (c *Client) getOrCreatePlaylist(name string) (string, error) {
	// Try to get existing playlist
	playlistID, err := c.getPlaylistIDByName(name)
	if err == nil {
		return playlistID, nil
	}

	// Create new playlist
	endpoint := "/Playlists"
	body := map[string]interface{}{
		"Name":          name,
		"MediaType":     "Video",
		"UserId":        "",
	}

	var response struct {
		ID string `json:"Id"`
	}

	if err := c.post(endpoint, body, &response); err != nil {
		return "", err
	}

	return response.ID, nil
}

func (c *Client) getItemIndexInPlaylist(itemID, playlistID string) (int, error) {
	endpoint := fmt.Sprintf("/Playlists/%s/Items", url.QueryEscape(playlistID))
	var response struct {
		Items []struct {
			ID       string `json:"Id"`
			PlaylistItemID string `json:"PlaylistItemId"`
		} `json:"Items"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return -1, err
	}

	for i, item := range response.Items {
		if item.ID == itemID {
			return i, nil
		}
	}

	return -1, fmt.Errorf("item not found in playlist")
}

func (c *Client) getTags(itemID string) ([]string, error) {
	endpoint := fmt.Sprintf("/Items/%s", url.QueryEscape(itemID))
	var response struct {
		TagItems []string `json:"TagItems"`
	}

	if err := c.get(endpoint, &response); err != nil {
		return nil, err
	}

	return response.TagItems, nil
}

func (c *Client) updateTags(itemID string, tags []string) error {
	endpoint := fmt.Sprintf("/Items/%s", url.QueryEscape(itemID))
	body := map[string]interface{}{
		"Id":       itemID,
		"TagItems": tags,
	}

	return c.post(endpoint, body, nil)
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
	req.Header.Set("X-Emby-Token", c.apiKey)
	
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
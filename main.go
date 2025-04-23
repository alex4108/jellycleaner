package main

import (
	"log"
	"os"
	"time"

	"github.com/alex4108/jellycleaner/config"
	"github.com/alex4108/jellycleaner/internal/jellyfin"
	"github.com/alex4108/jellycleaner/internal/radarr"
	"github.com/alex4108/jellycleaner/internal/sonarr"
)

const ( 
	expireTagConst = "Jellycleaner-Expire-"
)

func main() {
	log.Info("Starting jellycleaner")

	// Load configuration
	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize clients
	jellyfinClient, err := jellyfin.NewClient(cfg.Jellyfin.URL, os.Getenv("JELLYFIN_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to initialize Jellyfin client: %v", err)
	}

	sonarrClient, err := sonarr.NewClient(cfg.Sonarr.URL, os.Getenv("SONARR_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to initialize Sonarr client: %v", err)
	}

	radarrClient, err := radarr.NewClient(cfg.Radarr.URL, os.Getenv("RADARR_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to initialize Radarr client: %v", err)
	}

	jellyseerrClient, err := jellyseerr.NewClient(cfg.Jellyseerr.URL, os.Getenv("JELLYSEERR_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to initialize Jellyseerr client: %v", err)
	}

	
	processContent(cfg, jellyfinClient, sonarrClient, radarrClient, jellyseerrClient)

	log.Info("Job completed!")
}

func processContent(cfg *config.Config, jellyfinClient *jellyfin.Client, sonarrClient *sonarr.Client, radarrClient *radarr.Client) {
	log.Info("Starting content evaluation process...")

	// Process each library
	for _, library := range cfg.Jellyfin.Libraries {
		log.Infof("Processing library: %s", library.Name)

		// Get all items in the library
		items, err := jellyfinClient.GetLibraryItems(library.Name)
		if err != nil {
			log.Infof("Error getting library items for %s: %v", library.Name, err)
			continue
		}

		// Process each item
		for _, item := range items {
			// Skip if the item is in the exclusion list
			if isExcluded(item.Name, library.Exclusions) {
				log.Infof("Skipping excluded item: %s", item.Name)
				continue
			}

			// Check if item should be marked for deletion
			shouldDelete, reason := shouldMarkForDeletion(item, library, jellyfinClient)
			if shouldDelete {
				log.Infof("Marking item for deletion: %s (Reason: %s)", item.Name, reason)
				
				// Add to "Headed Out" playlist if not already there
				if !jellyfinClient.IsInPlaylist(item.ID, cfg.HeadedOutPlaylist.Name) {
					expirationDate := time.Now().AddDate(0, 0, cfg.HeadedOutPlaylist.DeletionDelayDays)
					if err := jellyfinClient.AddToPlaylist(item.ID, cfg.HeadedOutPlaylist.Name); err != nil {
						log.Infof("Failed to add %s to playlist: %v", item.Name, err)
					} else {
						// Add expiration tag
						tag := formatExpirationTag(expirationDate)
						if err := jellyfinClient.AddTag(item.ID, tag); err != nil {
							log.Infof("Failed to add expiration tag to %s: %v", item.Name, err)
						}
					}
				}
			} else {
				// If item is in playlist but shouldn't be, remove it
				if jellyfinClient.IsInPlaylist(item.ID, cfg.HeadedOutPlaylist.Name) {
					log.Infof("Removing item from deletion list: %s", item.Name)
					if err := jellyfinClient.RemoveFromPlaylist(item.ID, cfg.HeadedOutPlaylist.Name); err != nil {
						log.Infof("Failed to remove %s from playlist: %v", item.Name, err)
					}
					// Remove expiration tag
					expirationTags := jellyfinClient.GetExpirationTags(item.ID)
					for _, tag := range expirationTags {
						if err := jellyfinClient.RemoveTag(item.ID, tag); err != nil {
							log.Infof("Failed to remove expiration tag from %s: %v", item.Name, err)
						}
					}
				}
			}
		}
	}

	// Process items that are due for deletion
	processItemsDueForDeletion(cfg, jellyfinClient, sonarrClient, radarrClient)
}

func isExcluded(itemName string, exclusions []string) bool {
	for _, exclusion := range exclusions {
		if itemName == exclusion {
			return true
		}
	}
	return false
}

func shouldMarkForDeletion(item jellyfin.Item, library config.Library, jellyfinClient *jellyfin.Client) (bool, string) {
	// Check if the item has been watched by all users
	if library.Rules.DeleteIfWatchedByAll {
		watchedByAll, err := jellyfinClient.IsWatchedByAllUsers(item.ID)
		if err != nil {
			log.Infof("Error checking if %s is watched by all: %v", item.Name, err)
		} else if watchedByAll {
			return true, "Watched by all users"
		}
	}

	// Check if the item is older than the max age
	if library.Rules.MaxAgeDays > 0 {
		addedDate, err := jellyfinClient.GetItemAddedDate(item.ID)
		if err != nil {
			log.Infof("Error getting added date for %s: %v", item.Name, err)
		} else {
			ageInDays := int(time.Since(addedDate).Hours() / 24)
			if ageInDays > library.Rules.MaxAgeDays {
				return true, "Exceeds maximum age"
			}
		}
	}

	return false, ""
}

func formatExpirationTag(expirationDate time.Time) string {
	return expireTagConst + expirationDate.Format("2006-01-02")
}

func processItemsDueForDeletion(cfg *config.Config, jellyfinClient *jellyfin.Client, sonarrClient *sonarr.Client, radarrClient *radarr.Client) {
	log.Println("Processing items due for deletion...")

	// Get all items in the "Headed Out" playlist
	playlistItems, err := jellyfinClient.GetPlaylistItems(cfg.HeadedOutPlaylist.Name)
	if err != nil {
		log.Infof("Error getting playlist items: %v", err)
		return
	}

	now := time.Now()
	for _, item := range playlistItems {
		// Get expiration tag
		expirationTags := jellyfinClient.GetExpirationTags(item.ID)
		for _, tag := range expirationTags {
			// Parse expiration date from tag
			expDate, err := parseExpirationDate(tag)
			if err != nil {
				log.Infof("Error parsing expiration date for %s: %v", item.Name, err)
				continue
			}

			// Check if it's time to delete
			if now.After(expDate) {
				log.Infof("Deleting content: %s (Expiration: %s)", item.Name, expDate.Format("2006-01-02"))
				
				// Delete from Sonarr or Radarr first
				if item.Type == "Series" {
					if err := sonarrClient.DeleteSeries(item.ExternalID); err != nil {
						log.Errorf("Failed to delete series from Sonarr: %v", err)
						continue
					}
				} else if item.Type == "Movie" {
					if err := radarrClient.DeleteMovie(item.ExternalID); err != nil {
						log.Errorf("Failed to delete movie from Radarr: %v", err)
						continue
					}
				}

				// Remove from Jellyfin playlist and delete tags
				if err := jellyfinClient.RemoveFromPlaylist(item.ID, cfg.HeadedOutPlaylist.Name); err != nil {
					log.Errorf("Failed to remove %s from playlist: %v", item.Name, err)
				}
				if err := jellyfinClient.RemoveTag(item.ID, tag); err != nil {
					log.Errorf("Failed to remove expiration tag from %s: %v", item.Name, err)
				}
				
				// Try to remove it from Jellyseerr
				if err := jellyseerrClient.Remove(item); err != nil { 
					log.Warnf("Failed to remove content (%s) from Jellyseerr: %v", item.Name, err)
				}

				log.Infof("Successfully deleted: %s", item.Name)
			}
		}
	}
}

func parseExpirationDate(tag string) (time.Time, error) {
	
	dateStr := 
	return time.Parse("2006-01-02", dateStr)
}

func getConfigPath() string {
	// Check if config path is set via environment variable
	configPath := os.Getenv("JELLYCLEANER_CONFIG")
	if configPath != "" {
		return configPath
	}
	// Default to config.yaml in current directory
	return "config.yaml"
}
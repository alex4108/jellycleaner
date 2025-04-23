package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the top-level configuration
type Config struct {
	Jellyfin          JellyfinConfig     `yaml:"jellyfin"`
	Sonarr            SonarrConfig       `yaml:"sonarr"`
	Radarr            RadarrConfig       `yaml:"radarr"`
	Jellyseerr		 JellyseerrConfig    `yaml:"jellyseerr"`
	HeadedOutPlaylist PlaylistConfig     `yaml:"headed_out_playlist"`
}

// JellyfinConfig contains Jellyfin-specific configuration
type JellyfinConfig struct {
	URL       string     `yaml:"url"`
	Libraries []Library  `yaml:"libraries"`
}

type JellyseerrConfig struct {
	URL       string     `yaml:"url"`
}

// Library represents a single Jellyfin media library
type Library struct {
	Name       string        `yaml:"name"`
	Type       string        `yaml:"type"` // "movie" or "series"
	Rules      LibraryRules  `yaml:"rules"`
	Exclusions []string      `yaml:"exclusions"`
}

// LibraryRules defines conditions for marking content for deletion
type LibraryRules struct {
	DeleteIfWatchedByAll bool `yaml:"delete_if_watched_by_all"`
	MaxAgeDays           int  `yaml:"max_age_days"`
}

// SonarrConfig contains Sonarr-specific configuration
type SonarrConfig struct {
	URL    string `yaml:"url"`
}

// RadarrConfig contains Radarr-specific configuration
type RadarrConfig struct {
	URL    string `yaml:"url"`
}

// PlaylistConfig contains settings for the "Headed Out" playlist
type PlaylistConfig struct {
	Name              string `yaml:"name"`
	CheckIntervalHours int    `yaml:"check_interval_hours"`
	DeletionDelayDays int    `yaml:"deletion_delay_days"`
}

// LoadConfig reads and parses the configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	// Basic validation
	if config.Jellyfin.URL == "" {
		return fmt.Errorf("jellyfin URL is required")
	}
	if config.Sonarr.URL == "" {
		return fmt.Errorf("sonarr URL is required")
	}
	if config.Radarr.URL == "" {
		return fmt.Errorf("radarr URL is required")
	}
	if config.Jellyseerr.URL == "" {
		return fmt.Errorf("jellyseerr URL is required")
	}
	if config.HeadedOutPlaylist.Name == "" {
		config.HeadedOutPlaylist.Name = "Headed Out" // Set default
	}
	if config.HeadedOutPlaylist.CheckIntervalHours == 0 {
		config.HeadedOutPlaylist.CheckIntervalHours = 24 // Set default
	}
	if config.HeadedOutPlaylist.DeletionDelayDays == 0 {
		config.HeadedOutPlaylist.DeletionDelayDays = 7 // Set default
	}

	return nil
}
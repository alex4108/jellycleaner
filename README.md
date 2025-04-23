# jellycleaner
An app to manage content lifecycle between Jellyfin, Jellyseerr, Sonarr, Radarr

### Environment Variables

| Variable Name         | Description                              | Default Value       | Required |
|-----------------------|------------------------------------------|---------------------|----------|
| `JELLYCLEANER_CONFIG` | Path to the configuration YAML file.     | `config.yaml`       | No       |
| `RADARR_API_KEY`      | If Radarr is configured, the API Key.    | `None`              | No       |
| `SONARR_API_KEY`      | If Sonarr is configured, the API Key.    | `None`              | No       |
| `JELLYSEERR_API_KEY`  | If Jellyseerr is configured, the API Key.| `None`              | No       |
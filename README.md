# jellycleaner

An app to manage content lifecycle between Jellyfin, Jellyseerr, Sonarr, Radarr

### Build Status

![Tests (main)](https://img.shields.io/github/actions/workflow/status/alex4108/jellycleaner/test.yml?logo=go
)

![Build and Push](https://img.shields.io/github/actions/workflow/status/alex4108/jellycleaner/build-and-push.yml?logo=docker&label=Release)

![CodeQL](https://img.shields.io/github/actions/workflow/status/alex4108/jellycleaner/codeql.yml?logo=qualys&label=CodeQL
)

### Releases

Check releases on the right for the latest tag to use.

Production-ready images are always available at `ghcr.io/alex4108/jellycleanerr:{release_version}`

### Environment Variables

| Variable Name         | Description                              | Default Value       | Required |
|-----------------------|------------------------------------------|---------------------|----------|
| `JELLYCLEANER_CONFIG` | Path to the configuration YAML file.     | `config.yaml`       | No       |
| `RADARR_API_KEY`      | If Radarr is configured, the API Key.    | `None`              | No       |
| `SONARR_API_KEY`      | If Sonarr is configured, the API Key.    | `None`              | No       |
| `JELLYSEERR_API_KEY`  | If Jellyseerr is configured, the API Key.| `None`              | No       |
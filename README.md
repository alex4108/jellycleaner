# jellycleaner

An app to manage content lifecycle between Jellyfin, Jellyseerr, Sonarr, Radarr

### Build Status

![Tests (main)](https://img.shields.io/github/actions/workflow/status/alex/jellycleaner/workflows/test.yml?branch=main&label=Tests&logo=github)

![Build and Push](https://img.shields.io/github/actions/workflow/status/alex/jellycleaner/workflows/build-and-push.yml?branch=main&label=Build%20and%20Push&logo=docker)

![CodeQL](https://img.shields.io/github/actions/workflow/status/alex/jellycleaner/workflows/codeql.yml?branch=main&label=CodeQL&logo=github)

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
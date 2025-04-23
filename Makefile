# Variables
APP_NAME := jellyseerr
VERSION := 0.1.0
USERNAME := alex4108
GIT_REPO := $(shell basename -s .git `git config --get remote.origin.url`)
GHCR_REGISTRY := ghcr.io
IMAGE_NAME := $(GHCR_REGISTRY)/$(USERNAME)/$(APP_NAME)
TAG := $(VERSION)

# Docker-related commands
.PHONY: build
build: build-local
	@echo "Building Docker image: $(IMAGE_NAME):$(TAG)"
	docker build -t $(IMAGE_NAME):$(TAG) .
	docker tag $(IMAGE_NAME):$(TAG) $(IMAGE_NAME):latest

.PHONY: push
push: build
	@echo "Pushing Docker image to GHCR: $(IMAGE_NAME):$(TAG)"
	docker push $(IMAGE_NAME):$(TAG)
	docker push $(IMAGE_NAME):latest

.PHONY: run
run: build-local build
	@echo "Running Docker container from image: $(IMAGE_NAME):$(TAG)"
	docker run -v $(PWD)/config:/home/appuser/config --name $(APP_NAME) $(IMAGE_NAME):$(TAG)

.PHONY: run-detached
run-detached: build-local build
	@echo "Running Docker container in detached mode from image: $(IMAGE_NAME):$(TAG)"
	docker run -d -v $(PWD)/config:/home/appuser/config --name $(APP_NAME) $(IMAGE_NAME):$(TAG)

.PHONY: stop
stop:
	@echo "Stopping Docker container: $(APP_NAME)"
	docker stop $(APP_NAME)

.PHONY: clean
clean:
	@echo "Removing Docker container: $(APP_NAME)"
	docker rm -f $(APP_NAME) 2>/dev/null || true

# Development-related commands
.PHONY: deps
deps:
	@echo "Downloading Go dependencies"
	go mod download

.PHONY: test
test:
	@echo "Running tests"
	go test -v ./...

.PHONY: build-local
build-local:
	@echo "Building binary locally"
	mkdir -p ./run
	go build -o ./run/$(APP_NAME) .

.PHONY: run-local
run-local: build-local
	@echo "Running binary locally"
	JELLYCLEANER_CONFIG=./config.yaml ./$(APP_NAME)

# GitHub Container Registry login helper
.PHONY: ghcr-login
ghcr-login:
	@echo "Logging in to GitHub Container Registry"
	@docker login $(GHCR_REGISTRY) -u $(USERNAME) -p $(shell cat .gh_pat)

# Help command
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build         - Build Docker image"
	@echo "  make push          - Push Docker image to GHCR"
	@echo "  make run           - Run Docker container"
	@echo "  make run-detached  - Run Docker container in detached mode"
	@echo "  make stop          - Stop Docker container"
	@echo "  make clean         - Remove Docker container"
	@echo "  make deps          - Download Go dependencies"
	@echo "  make test          - Run tests"
	@echo "  make build-local   - Build binary locally"
	@echo "  make run-local     - Run binary locally"
	@echo "  make ghcr-login    - Login to GitHub Container Registry"
	@echo "  make help          - Show this help message"
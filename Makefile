PACKAGE				:= asciidoctor-watcher
OWNER				:= thofisch
REPO				:= watcher
PROJECT				:= github.com/$(OWNER)/$(REPO)
DOCKER_IMAGE_NAME	= $(OWNER)/$(PACKAGE):$(VERSION)
DATE				?= $(shell date +%FT%T%z)
VERSION				?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || cat $(CURDIR)/.version 2> /dev/null || echo v0)
COMMIT				?= $(shell git rev-parse HEAD 2> /dev/null)
#CONFIG				= $(shell go list ./internal/config)
#LDFLAGS 			= -X $(CONFIG).Version=$(VERSION) -X $(CONFIG).BuildDate=$(DATE) -X $(CONFIG).Commit=$(COMMIT)
#ALL_PLATFORMS		:= linux/amd64 darwin/amd64 windows/amd64
#CURRENT_OS			= $(shell go env GOOS)
#CURRENT_ARCH		= $(shell go env GOARCH)
#OS					:= $(if $(GOOS),$(GOOS),$(CURRENT_OS))
#ARCH				:= $(if $(GOARCH),$(GOARCH),$(CURRENT_ARCH))
BIN					:= $(CURDIR)/bin
EXECUTABLE			:= watcher
OUTBIN				:= $(BIN)/$(EXECUTABLE)
M					= $(shell printf "\033[34;1m▶\033[0m")

export GO111MODULE=on

.PHONY: all
all: build

build: ; $(info $(M) Building binary $(OUTBIN)) @ ## build binary for current OS architecture
	@go build -o $(OUTBIN)

release: ; $(info $(M) Building release binary $(OUTBIN)) @ ## build release linux binary
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o $(OUTBIN)
#@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(OUTBIN) ./cmd/mystico

.PHONY: docker
docker: docker-build docker-push ## build and push docker container

.PHONY: docker-build
docker-build: ; $(info $(M) Building docker container $(DOCKER_IMAGE_NAME)) @ ## build docker image
	@docker build --build-arg LDFLAGS="$(LDFLAGS)" -t $(DOCKER_IMAGE_NAME) .

.PHONY: docker-push
docker-push: ; $(info $(M) Pushing docker container $(DOCKER_IMAGE_NAME)) @ ## push docker image
	@docker push $(DOCKER_IMAGE_NAME)

.PHONY: clean
clean: ; $(info $(M) Cleaning...) @ ## clean the build artifacts
	@rm -rf $(BIN)

.PHONY: version
version: ## prints the version (from either environment VERSION, git describe, or .version. default: v0)
	@echo $(VERSION)

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-19s\033[0m %s\n", $$1, $$2}'

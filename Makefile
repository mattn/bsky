BIN := bsky
ifeq ($(OS),Windows_NT)
BIN := $(BIN).exe
endif
VERSION := $$(make -s show-version)
CURRENT_REVISION := $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS := "-s -w -X main.revision=$(CURRENT_REVISION)"
GOOS := $(shell go env GOOS)
GOBIN ?= $(shell go env GOPATH)/bin
export GO111MODULE=on

GIT_SHA := $(shell git rev-parse HEAD)
GIT_SHA_SHORT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION := $(shell git describe --tags)-$(GIT_SHA_SHORT)
LDFLAGS := -s -w \
        -X 'github.com/jlewi/bsctl/pkg.Date=$(DATE)' \
        -X 'github.com/jlewi/bsctl/pkg.Version=$(subst v,,$(VERSION))' \
        -X 'github.com/jlewi/bsctl/pkg.Commit=$(GIT_SHA)' \
		-X 'github.com/jlewi/bsctl/pkg.BuiltBy=$(GIT_SHA)'

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) .

.PHONY: pwa
pwa:
	GOOS=js GOARCH=wasm go build -o web/app.wasm -ldflags="$(LDFLAGS)" ./pwa
	go build -o .build/pwa-server -ldflags="$(LDFLAGS)" ./pwa

# Build a static website
.PHONY: pwa
static:	
	GOOS=js GOARCH=wasm go build -o .build/static/web/app.wasm -ldflags="$(LDFLAGS)" ./pwa
	BUILD_STATIC=.build/static go run ./pwa

.PHONY: release
release:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) .
	zip -r bsky-$(GOOS)-$(VERSION).zip $(BIN)

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) .

.PHONY: show-version
show-version: $(GOBIN)/gobump
	gobump show -r .

$(GOBIN)/gobump:
	go install github.com/x-motemen/gobump/cmd/gobump@latest

.PHONY: test
test: build
	go test -v ./...

.PHONY: clean
clean:
	go clean

.PHONY: bump
bump: $(GOBIN)/gobump
ifneq ($(shell git status --porcelain),)
	$(error git workspace is dirty)
endif
ifneq ($(shell git rev-parse --abbrev-ref HEAD),main)
	$(error current branch is not main)
endif
	@gobump up -w .
	git commit -am "bump up version to $(VERSION)"
	git tag "v$(VERSION)"
	git push origin main
	git push origin "refs/tags/v$(VERSION)"

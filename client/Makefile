SHELL=/bin/bash

BINARY_NAME=opencode-piko
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

GO=go
CGO_ENABLED=0
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -s -w"

DIST_DIR=dist
EMBED_DIR=src/embed
OPENCODE_VERSION?=latest

GITHUB_REPO=anomalyco/opencode

PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

.PHONY: all
all: build

.PHONY: build
build: download-opencode
	@echo "Building ${BINARY_NAME} for $$(go env GOOS)/$$(go env GOARCH)..."
	@mkdir -p ${DIST_DIR}
	CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME} ./main.go
	@echo "Done: ${DIST_DIR}/${BINARY_NAME}"

.PHONY: build-all
build-all:
	@echo "Building all platforms..."
	@mkdir -p ${DIST_DIR}
	@for platform in ${PLATFORMS}; do \
		IFS='/' read -r os arch <<< "$$platform"; \
		echo "=== Building $$os/$$arch ==="; \
		$(MAKE) _build-single TARGET_OS=$$os TARGET_ARCH=$$arch; \
	done
	@echo "All builds complete: ${DIST_DIR}/"
	@ls -lh ${DIST_DIR}/

.PHONY: _build-single
_build-single:
	@$(MAKE) download-opencode TARGET_OS=$(TARGET_OS) TARGET_ARCH=$(TARGET_ARCH)
	@suffix=""; \
	if [ "$(TARGET_OS)" = "windows" ]; then suffix=".exe"; fi; \
	echo "Compiling $(TARGET_OS)/$(TARGET_ARCH)..."; \
	CGO_ENABLED=0 GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) ${GO} build ${LDFLAGS} \
		-o ${DIST_DIR}/${BINARY_NAME}-$(TARGET_OS)-$(TARGET_ARCH)$$suffix ./main.go

.PHONY: download-opencode
download-opencode:
	@mkdir -p ${EMBED_DIR}
	@target_os=$${TARGET_OS:-$$(go env GOOS)}; \
	target_arch=$${TARGET_ARCH:-$$(go env GOARCH)}; \
	oc_arch=$$target_arch; \
	if [ "$$oc_arch" = "amd64" ]; then oc_arch="x64"; fi; \
	ext=".tar.gz"; \
	if [ "$$target_os" = "darwin" ] || [ "$$target_os" = "windows" ]; then ext=".zip"; fi; \
	filename="opencode-$$target_os-$$oc_arch$$ext"; \
	if [ "${OPENCODE_VERSION}" = "latest" ]; then \
		url="https://github.com/${GITHUB_REPO}/releases/latest/download/$$filename"; \
	else \
		url="https://github.com/${GITHUB_REPO}/releases/download/v${OPENCODE_VERSION}/$$filename"; \
	fi; \
	echo "Downloading $$filename..."; \
	tmpdir=$$(mktemp -d); \
	curl -sL -o "$$tmpdir/$$filename" "$$url"; \
	filesize=$$(wc -c < "$$tmpdir/$$filename" | tr -d ' '); \
	if [ "$$filesize" -lt 1000 ]; then echo "Download failed: $$url"; cat "$$tmpdir/$$filename"; exit 1; fi; \
	echo "Extracting opencode binary..."; \
	if [ "$$ext" = ".zip" ]; then \
		unzip -qo "$$tmpdir/$$filename" -d "$$tmpdir/extract"; \
		bin_name="opencode"; \
		if [ "$$target_os" = "windows" ]; then bin_name="opencode.exe"; fi; \
		find "$$tmpdir/extract" -name "$$bin_name" -type f | head -1 | xargs -I{} cp {} "$$tmpdir/opencode_bin"; \
	else \
		mkdir -p "$$tmpdir/extract"; \
		tar -xzf "$$tmpdir/$$filename" -C "$$tmpdir/extract"; \
		find "$$tmpdir/extract" -name "opencode" -type f | head -1 | xargs -I{} cp {} "$$tmpdir/opencode_bin"; \
	fi; \
	echo "Compressing for embed..."; \
	gzip -9 -c "$$tmpdir/opencode_bin" > ${EMBED_DIR}/opencode.gz; \
	rm -rf "$$tmpdir"; \
	echo "Embedded: ${EMBED_DIR}/opencode.gz ($$(du -h ${EMBED_DIR}/opencode.gz | cut -f1))"

.PHONY: docker
docker: download-opencode
	@target_os=$${TARGET_OS:-linux}; \
	target_arch=$${TARGET_ARCH:-amd64}; \
	echo "Building Docker image for $$target_os/$$target_arch..."; \
	CGO_ENABLED=0 GOOS=$$target_os GOARCH=$$target_arch ${GO} build ${LDFLAGS} \
		-o ${DIST_DIR}/${BINARY_NAME}-$$target_os-$$target_arch ./main.go; \
	docker build --platform $$target_os/$$target_arch \
		--build-arg BINARY=${DIST_DIR}/${BINARY_NAME}-$$target_os-$$target_arch \
		--build-arg http_proxy="" --build-arg https_proxy="" \
		--build-arg HTTP_PROXY="" --build-arg HTTPS_PROXY="" \
		-t opencode-piko:latest .; \
	echo "Docker image: opencode-piko:latest"

.PHONY: clean
clean:
	@rm -rf ${DIST_DIR}
	@rm -f ${EMBED_DIR}/opencode.gz
	@echo "Cleaned"

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build              - Build for current platform"
	@echo "  build-all          - Build for all 6 platforms"
	@echo "  docker             - Build Docker image (linux/amd64 default)"
	@echo "  download-opencode  - Download opencode binary for embedding"
	@echo "  clean              - Remove build artifacts"
	@echo ""
	@echo "Variables:"
	@echo "  OPENCODE_VERSION   - Version to download (default: latest)"
	@echo "  TARGET_OS          - Target OS for download"
	@echo "  TARGET_ARCH        - Target arch for download"

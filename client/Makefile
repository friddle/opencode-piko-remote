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

OPENCODE_SRC=../opencode
OPENCODE_DIST=$(OPENCODE_SRC)/packages/opencode/dist

PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

.PHONY: all
all: build

.PHONY: build
build: download-opencode
	@echo "Building ${BINARY_NAME} for $$(go env GOOS)/$$(go env GOARCH)..."
	@mkdir -p ${DIST_DIR}
	CGO_ENABLED=0 ${GO} build ${LDFLAGS} -o ${DIST_DIR}/${BINARY_NAME} .
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
		-o ${DIST_DIR}/${BINARY_NAME}-$(TARGET_OS)-$(TARGET_ARCH)$$suffix .

.PHONY: build-opencode
build-opencode:
	@if [ ! -f "$(OPENCODE_DIST)/opencode-darwin-arm64/bin/opencode" ]; then \
		echo "Building opencode from submodule (all platforms)..."; \
		cd $(OPENCODE_SRC) && bun install; \
		cd packages/opencode && bun run build -- --skip-install; \
	else \
		echo "opencode already built, skipping"; \
	fi

.PHONY: download-opencode
download-opencode: build-opencode
	@mkdir -p ${EMBED_DIR}
	@target_os=$${TARGET_OS:-$$(go env GOOS)}; \
	target_arch=$${TARGET_ARCH:-$$(go env GOARCH)}; \
	oc_arch=$$target_arch; \
	if [ "$$oc_arch" = "amd64" ]; then oc_arch="x64"; fi; \
	bin_path="$(OPENCODE_DIST)/opencode-$$target_os-$$oc_arch/bin/opencode"; \
	if [ ! -f "$$bin_path" ]; then echo "Binary not found: $$bin_path"; exit 1; fi; \
	echo "Embedding opencode from: $$bin_path"; \
	gzip -9 -c "$$bin_path" > ${EMBED_DIR}/opencode.gz; \
	echo "Embedded: ${EMBED_DIR}/opencode.gz ($$(du -h ${EMBED_DIR}/opencode.gz | cut -f1))"

.PHONY: docker
docker: download-opencode
	@target_os=$${TARGET_OS:-linux}; \
	target_arch=$${TARGET_ARCH:-amd64}; \
	echo "Building Docker image for $$target_os/$$target_arch..."; \
	CGO_ENABLED=0 GOOS=$$target_os GOARCH=$$target_arch ${GO} build ${LDFLAGS} \
		-o ${DIST_DIR}/${BINARY_NAME}-$$target_os-$$target_arch .; \
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
	@echo "  build-opencode     - Build opencode from submodule (all platforms)"
	@echo "  download-opencode  - Embed opencode binary for current/target platform"
	@echo "  clean              - Remove build artifacts"
	@echo ""
	@echo "Variables:"
	@echo "  TARGET_OS          - Target OS for embed"
	@echo "  TARGET_ARCH        - Target arch for embed"

# Makefile for opencode-piko-server

# 变量定义
IMAGE_NAME=opencode-piko-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker相关变量
DOCKER_REGISTRY?=ghcr.io
DOCKER_NAMESPACE?=friddle
DOCKER_TAG?=latest

# 构建目录
BUILD_DIR=dist
DIST_DIR=dist

# 默认目标
.PHONY: all
all: clean build

# 构建Docker镜像
.PHONY: build
build:
	@echo "构建 ${IMAGE_NAME} Docker镜像..."
	@mkdir -p ${BUILD_DIR}
	docker build -f build/Dockerfile -t ${IMAGE_NAME}:${DOCKER_TAG} .
	@echo "Docker镜像构建完成: ${IMAGE_NAME}:${DOCKER_TAG}"

# 构建并推送Docker镜像
.PHONY: build-push
build-push: build
	@echo "推送Docker镜像到注册表..."
	@if [ -n "${DOCKER_REGISTRY}" ]; then \
		docker tag ${IMAGE_NAME}:${DOCKER_TAG} ${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${IMAGE_NAME}:${DOCKER_TAG}; \
		docker push ${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${IMAGE_NAME}:${DOCKER_TAG}; \
		echo "镜像已推送到: ${DOCKER_REGISTRY}/${DOCKER_NAMESPACE}/${IMAGE_NAME}:${DOCKER_TAG}"; \
	else \
		echo "DOCKER_REGISTRY 未设置，跳过推送"; \
	fi

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	@rm -rf ${BUILD_DIR} ${DIST_DIR}
	@docker rmi ${IMAGE_NAME}:${DOCKER_TAG} 2>/dev/null || true
	@echo "清理完成"

# 帮助
.PHONY: help
help:
	@echo "可用的 Make 目标:"
	@echo "  build         - 构建Docker镜像"
	@echo "  build-push    - 构建并推送Docker镜像"
	@echo "  help          - 显示此帮助信息" 
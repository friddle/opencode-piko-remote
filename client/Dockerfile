ARG REGISTRY=docker.io
FROM ${REGISTRY}/library/debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

ARG BINARY=dist/opencode-piko-linux-amd64
COPY ${BINARY} /usr/local/bin/opencode-piko

RUN useradd -m -s /bin/bash opencode
USER opencode
WORKDIR /home/opencode

ENTRYPOINT ["opencode-piko"]

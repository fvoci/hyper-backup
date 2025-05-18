# 📄 Dockerfile

#── Builder stage ─────────────────────────────────────────────────────────────
FROM golang:1.24-bookworm AS builder
ARG TARGETOS TARGETARCH

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -o hyper-backup main.go

# ── Final stage ───────────────────────────────────────────────────────────────
FROM debian:bookworm AS final
ARG MONGO_TOOLS_VERSION=100.12.0
ARG UID=1001
ARG GID=1001
ARG TZ=Asia/Seoul

# System user setup
RUN set -eux; \
    groupadd --gid ${GID} hyper-backup; \
    useradd --uid ${UID} --gid ${GID} --home-dir /home/hyper-backup --create-home hyper-backup

ENV TZ=${TZ}
WORKDIR /home/hyper-backup

# Dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates curl wget gnupg lsb-release \
    rsync rclone default-mysql-client postgresql-client \
 && apt-get clean && rm -rf /var/lib/apt/lists/*

# MongoDB Tools install
RUN set -eux; \
    dpkgArch="$(dpkg --print-architecture | awk -F- '{ print $NF }')"; \
    case "${dpkgArch}" in \
      amd64) ARCH_TAG="x86_64";; \
      arm64) ARCH_TAG="arm64";; \
      *) echo "Unsupported arch ${dpkgArch}"; exit 1;; \
    esac; \
    URL="https://fastdl.mongodb.org/tools/db/mongodb-database-tools-ubuntu2404-${ARCH_TAG}-${MONGO_TOOLS_VERSION}.tgz"; \
    TMPDIR="$(mktemp -d)"; \
    curl -fsSL "${URL}" | tar -xz -C "${TMPDIR}"; \
    mv "${TMPDIR}"/mongodb-database-tools-*/bin/* /usr/local/bin/; \
    rm -rf "${TMPDIR}"

# Copy our hyper-backup binary into PATH
COPY --from=builder /app/hyper-backup /usr/bin/hyper-backup
RUN chmod +x /usr/bin/hyper-backup

USER hyper-backup

ENTRYPOINT ["hyper-backup"]


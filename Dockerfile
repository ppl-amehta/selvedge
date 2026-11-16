# ========================================================
# FPCC SELVEDGE ENTERPRISE PRODUCTION CONTAINER
# Multi-Stage Zero-CVE Hardened Alpine Build
# Target Architecture: linux/amd64, linux/arm64
# ========================================================

# --------------------------------------------------------
# Stage 1: Build Svelte 5 Embedded Frontend (Native Builder Platform)
# --------------------------------------------------------
FROM --platform=$BUILDPLATFORM node:24-alpine AS web-builder

WORKDIR /build/web

# Cache package manifests first
COPY web/package*.json ./
RUN npm ci --no-audit --prefer-offline

# Copy frontend source and build optimized dist
COPY web/ ./
RUN npm run build

# --------------------------------------------------------
# Stage 2: Statically Compile Go 1.27 + CGO SQLite Binary
# --------------------------------------------------------
FROM golang:alpine AS go-builder

RUN apk add --no-cache musl-dev gcc git ca-certificates tzdata

WORKDIR /build

# Cache Go module dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree and embed built web distribution
COPY . .
COPY --from=web-builder /build/web/dist ./web/dist

# Build fully static CGO binary without glibc dependencies
RUN CGO_ENABLED=1 GOOS=linux \
    go build -tags "netgo,osusergo" \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/selvedge ./cmd/selvedge

# Strictly verify binary is 100% static (matches both musl and glibc static output)
RUN ldd /bin/selvedge 2>&1 | grep -iE "not a (dynamic executable|valid dynamic program)"

# --------------------------------------------------------
# Stage 3: Minimal Hardened Alpine 3.21 Runtime (~25MB)
# --------------------------------------------------------
FROM alpine:3.21

# Install CA certs, timezone data, and lightweight probe utility
RUN apk add --no-cache ca-certificates tzdata wget \
    && rm -rf /var/cache/apk/*

# Create dedicated non-root application user (UID 10001)
RUN addgroup -g 10001 -S appgroup \
    && adduser -u 10001 -S appuser -G appgroup

# Create persistent storage mount directory owned by appuser
RUN mkdir -p /data && chown -R appuser:appgroup /data

# Copy binary from builder with strict permissions
COPY --from=go-builder --chown=appuser:appgroup /bin/selvedge /usr/local/bin/selvedge

# Runtime environment settings
ENV PORT=4000 \
    DB_PATH=/data/selvedge.db \
    TZ=UTC

USER appuser
WORKDIR /home/appuser

VOLUME ["/data"]
EXPOSE 4000

# High-frequency production healthcheck
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --spider -q http://127.0.0.1:4000/api/health || exit 1

ENTRYPOINT ["/usr/local/bin/selvedge"]
CMD ["-db", "/data/selvedge.db", "-port", "4000"]

# --- Stage 1: Build ---
FROM --platform=$BUILDPLATFORM golang:1.25.5-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Leverage Docker layer caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build static binaries
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o relay ./cmd/relay/main.go && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o cli ./cmd/cli

# --- Stage 2: Final Runtime ---
# The official Google Distroless static image for Go
FROM gcr.io/distroless/static-debian12:nonroot

LABEL maintainer="Open Outbox Team"

WORKDIR /app

# Copy binaries from builder using the built-in nonroot user (UID 65532)
COPY --from=builder --chown=nonroot:nonroot /app/relay .
COPY --from=builder --chown=nonroot:nonroot /app/cli .

# Standard port for health checks and metrics
EXPOSE 8080

# Use the standard Google nonroot user
USER nonroot:nonroot

# Execute the binary directly
ENTRYPOINT ["/app/relay"]

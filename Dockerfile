# Stage 1: Build
FROM golang:1.24.5 AS builder

WORKDIR /app

# Copy dependency manifests and download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build all services at once
RUN for svc in api scheduler worker; do \
      go build -o /bin/aster-$svc ./cmd/aster-$svc; \
    done


# Stage 2: Runtime
FROM debian:bookworm-slim

WORKDIR /app

# Copy all built binaries from builder
COPY --from=builder /bin/aster-* /bin/

# Default entrypoint (each service overrides with `command:` in compose)
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["echo 'Please specify a service binary to run (aster-api, aster-scheduler, aster-worker)'"]

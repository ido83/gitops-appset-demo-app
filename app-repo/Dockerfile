# syntax=docker/dockerfile:1.6
# Multi-stage build:
# 1) Build a static Go binary
# 2) Copy into a minimal, non-root runtime image

FROM golang:1.22-alpine AS build

# Install CA certs for any HTTPS calls during build (e.g., go mod download).
RUN apk add --no-cache ca-certificates git

WORKDIR /src

# Copy module files first for better Docker layer caching.
COPY go.mod go.sum ./
RUN go mod download || true

# Copy the rest of the source.
COPY . .

# Build metadata injected by CI (Jenkins).
ARG VERSION=0.0.0
ARG GIT_SHA=dev
ARG BUILD_TIME=unknown

# Build a static binary (CGO_DISABLED) so we can use distroless static runtime.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
    -ldflags "-s -w \
    -X main.Version=${VERSION} \
    -X main.GitSHA=${GIT_SHA} \
    -X main.BuildTime=${BUILD_TIME}" \
    -o /out/hello-web ./cmd/hello-web

# Distroless static runtime, nonroot user by default — production best practice.
FROM gcr.io/distroless/static-debian12:nonroot

# The app listens on 8080 by default (configurable with PORT env var).
EXPOSE 8080

# Copy binary.
COPY --from=build /out/hello-web /hello-web

# Non-root execution is an important hardening step.
USER nonroot:nonroot

ENTRYPOINT ["/hello-web"]

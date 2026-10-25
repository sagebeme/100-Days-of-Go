# syntax=docker/dockerfile:1
#
# Build from the repo root, so the build can see go.mod and go.sum:
#   docker build -f phase-5-backend/day-063-docker-and-ci/solution/Dockerfile -t twende-api .
#   docker run --rm -p 8080:8080 twende-api

# Stage 1: build a static binary with the full Go toolchain.
FROM golang:1.26 AS build
WORKDIR /src

# Download modules first. This layer is cached until go.mod or go.sum changes,
# so editing code doesn't download everything again.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0: no C libraries, so the binary runs on an image with no libc.
# -trimpath drops local file paths; -ldflags="-s -w" drops debug symbols.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/api ./phase-5-backend/day-063-docker-and-ci/solution/cmd/api

# Stage 2: only the binary, on distroless: no shell, no package manager, CA
# certificates and time zone data included, and a nonroot user (UID 65532).
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
ENV PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD ["/api", "-healthcheck"]
ENTRYPOINT ["/api"]

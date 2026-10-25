# syntax=docker/dockerfile:1
#
# Build from the repo root, so the build can see go.mod and go.sum:
#   docker build -f phase-5-backend/day-063-docker-and-ci/starter/Dockerfile -t twende-api .
#
# Your code here. Follow step 7 in the README:
#
# Stage 1, "build", FROM golang:1.26
#   - WORKDIR /src
#   - COPY go.mod and go.sum, then RUN go mod download (a cached layer)
#   - COPY the rest
#   - RUN a static build (CGO_ENABLED=0) of ./phase-5-backend/day-063-docker-and-ci/starter/cmd/api
#     into /out/api, with -trimpath and -ldflags="-s -w"
#
# Stage 2, FROM gcr.io/distroless/static-debian12:nonroot
#   - COPY --from=build the binary to /api
#   - USER nonroot:nonroot
#   - ENV PORT=8080 and EXPOSE 8080
#   - HEALTHCHECK that runs ["/api", "-healthcheck"]
#   - ENTRYPOINT ["/api"]

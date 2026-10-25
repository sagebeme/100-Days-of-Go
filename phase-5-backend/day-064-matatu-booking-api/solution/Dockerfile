# syntax=docker/dockerfile:1
#
# From the repo root:
#   docker build -f phase-5-backend/day-064-matatu-booking-api/solution/Dockerfile -t matatu-api .
#   docker run --rm -p 8080:8080 -e JWT_SECRET="$(openssl rand -base64 32)" matatu-api
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/api ./phase-5-backend/day-064-matatu-booking-api/solution/cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/api"]

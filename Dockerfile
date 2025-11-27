# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /src

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy the remainder of the workspace
COPY . .

# Build service binaries
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/tracking-service ./cmd/tracking-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/tracking-worker ./cmd/tracking-worker

# Tracking service image
FROM gcr.io/distroless/base-debian12:nonroot AS tracking-service
COPY --from=builder /out/tracking-service /usr/local/bin/tracking-service
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/tracking-service"]

# Tracking worker image
FROM gcr.io/distroless/base-debian12:nonroot AS tracking-worker
COPY --from=builder /out/tracking-worker /usr/local/bin/tracking-worker
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/tracking-worker"]

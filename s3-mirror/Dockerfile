# Build stage
FROM golang:1.21 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 go build -o /build/k8s-binary-downloader main.go

FROM scratch

WORKDIR /app
COPY --from=builder /build/k8s-binary-downloader /app/k8s-binary-downloader

ENTRYPOINT ["/app/k8s-binary-downloader"]
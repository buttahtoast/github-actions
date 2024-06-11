FROM golang:1.21

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o k8s-binary-downloader main.go

ENTRYPOINT ["./k8s-binary-downloader"]
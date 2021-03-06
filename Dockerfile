FROM golang:1.13.1-alpine3.10 as builder
RUN mkdir /src
ADD . /src/
WORKDIR /src
RUN go mod download
RUN GOFLAGS=-mod=vendor go build -ldflags "-s -w -X main.version=$(cat VERSION)"
FROM alpine
COPY --from=builder /src/jaeger-bomb /app/jaeger-bomb
WORKDIR /app
ENTRYPOINT ["/app/jaeger-bomb"]
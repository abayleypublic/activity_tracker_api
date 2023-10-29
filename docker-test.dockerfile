FROM golang:1.20.3-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache musl-dev gcc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

CMD ["/bin/sh", "-c", "go test -coverprofile coverage.out -json ./... > ./report.json"]
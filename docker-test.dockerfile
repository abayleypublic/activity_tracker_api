FROM golang:1.23.10-alpine3.22

RUN apk update && apk upgrade && \
    apk add --no-cache musl-dev gcc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

CMD ["/bin/sh", "-c", "go test -coverprofile coverage.out -json ./... > ./report.json"]
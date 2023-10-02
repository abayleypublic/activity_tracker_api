FROM golang:1.20.3-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache musl-dev gcc

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", "air.toml"]
FROM golang:1.23.10-alpine3.22

RUN apk update && apk upgrade && \
    apk add --no-cache musl-dev gcc

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", "air.toml"]
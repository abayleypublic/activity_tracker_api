FROM golang:1.20.3-alpine

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest
# RUN go mod download

CMD ["air", "-c", "air.toml"]
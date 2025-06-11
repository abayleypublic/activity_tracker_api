FROM golang:1.20.3-alpine

ARG TARGETOS TARGETARCH TARGETVARIANT

RUN apk update && apk upgrade && \
    apk add --no-cache musl-dev gcc

WORKDIR /app
COPY . .
RUN go mod download
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} CGO_ENABLED=1 go build -o /activity-tracker-api ./cmd/server/main.go

ENTRYPOINT ["/activity-tracker-api"]
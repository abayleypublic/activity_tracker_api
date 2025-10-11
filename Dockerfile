FROM golang:1.25.2-alpine3.22

ARG TARGETOS TARGETARCH TARGETVARIANT

RUN apk update && apk upgrade && \
    apk add --no-cache musl-dev gcc

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} CGO_ENABLED=1 go build -o /activity-tracker-api ./cmd/server/main.go

ENTRYPOINT ["/activity-tracker-api"]
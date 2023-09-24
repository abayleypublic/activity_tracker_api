FROM golang:1.20.3-alpine as builder

WORKDIR /app

COPY . .
RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -o /activity-tracker-api ./cmd/server/main.go

FROM scratch

COPY --from=builder /activity-tracker-api /activity-tracker-api

ENTRYPOINT ["/activity-tracker-api"]
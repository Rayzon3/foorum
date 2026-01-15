FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.19.2
RUN go build -o /bin/server ./cmd/server

FROM alpine:3.20

RUN adduser -D -g '' app

WORKDIR /app

COPY --from=builder /bin/server /app/server
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/db /app/db
COPY docker-entrypoint.sh /app/docker-entrypoint.sh

RUN chmod +x /app/docker-entrypoint.sh

USER app

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]

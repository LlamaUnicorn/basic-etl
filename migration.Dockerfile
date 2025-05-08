FROM golang:1.21-alpine

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.14.0

COPY migrations ./migrations

CMD ["goose", "-dir", "migrations", "postgres", "host=pg port=5432 dbname=note user=note-user password=note-password sslmode=disable", "up"]
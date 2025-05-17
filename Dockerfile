FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /service

FROM alpine:3.20.6
COPY --from=builder /service /service
CMD ["/service"]
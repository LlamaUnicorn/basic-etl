FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /service

FROM alpine:3.19
COPY --from=builder /service /service
CMD ["/service"]
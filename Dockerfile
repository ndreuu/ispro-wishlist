FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /wishlist-service .

FROM alpine:latest

WORKDIR /

RUN apk add --no-cache ca-certificates

COPY --from=builder /wishlist-service /wishlist-service

EXPOSE 8080

CMD ["/wishlist-service"]

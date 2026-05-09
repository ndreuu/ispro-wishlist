FROM golang:1.23-alpine AS builder

WORKDIR /app

# Устанавливаем сертификаты
RUN apk add --no-cache ca-certificates && update-ca-certificates

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем модули с явным указанием proxy
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /wishlist-service .

# Финальный образ
FROM alpine:latest

WORKDIR /

COPY --from=builder /wishlist-service /wishlist-service
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

CMD ["/wishlist-service"]

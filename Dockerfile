FROM alpine:latest

WORKDIR /

COPY wishlist-service /wishlist-service

EXPOSE 8080

CMD ["/wishlist-service"]

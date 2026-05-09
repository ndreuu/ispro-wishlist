# Wishlist Service

## Описание

**Wishlist Service** – это REST API сервис для управления списками желаемых товаров.

Сервис позволяет пользователям:

- создавать вишлисты;
- просматривать список всех вишлистов;
- получать конкретный вишлист по идентификатору;
- удалять вишлисты;
- добавлять товары в вишлист;
- удалять товары из вишлиста.

## API Documentation

- [Swagger UI](https://ndreuu.github.io/ispro-wishlist/). Выполнение запросов возможно при запуске локального сервера.


## Сборка и запуск

```
go mod tidy
go run .
```

Сервер будет доступен по адресу:
```
http://localhost:8080
```

Swagger UI доступен по адресу:
```
http://localhost:8080/swagger/index.html
```

## Lab3 Monitoring

Сервис предоставляет Prometheus метрики и Grafana dashboard для мониторинга.

### Metrics Endpoint
```
http://localhost:8080/metrics
```

### Бизнес-метрики (custom product metrics)
- `wishlist_service_wishlists_created_total` - количество созданных вишлистов
- `wishlist_service_wishlists_get_total` - количество операций чтения вишлистов
- `wishlist_service_items_added_total` - количество добавленных элементов
- `wishlist_service_items_read_total` - количество элементов, возвращённых в ответах
- `wishlist_service_items_in_response` - распределение количества элементов в ответах API

### Примеры PromQL запросов для Grafana
```promql
# Скорость создания вишлистов
sum(rate(wishlist_service_wishlists_created_total[5m]))

# Скорость чтения вишлистов
sum(rate(wishlist_service_wishlists_get_total[5m]))

# Количество добавленных элементов за час
sum(increase(wishlist_service_items_added_total[1h]))

# Количество прочитанных элементов за час
sum(increase(wishlist_service_items_read_total[1h]))

# 95-й перцентиль количества элементов в ответе
histogram_quantile(0.95, sum(rate(wishlist_service_items_in_response_bucket[5m])) by (le))
```

![alt text](monitoring/image.png)
![alt text](monitoring/image-1.png)
![alt text](monitoring/image-2.png)
# Marketplace Architecture

### Домены и сервисы

| Домен | Сервис | Ответственность |
|-------|--------|-----------------|
| Управление пользователями | User Service | Регистрация, аутентификация, профили, роли |
| Товарный каталог | Catalog Service | CRUD товаров, категории, поиск |
| Заказы | Order Service | Создание заказов, статусы, история |
| Платежи | Payment Service | Интеграция с платёжными шлюзами, транзакции |
| Уведомления | Notification Service | Email, push-уведомления |
| Персонализация | Feed Service | Лента рекомендаций, предпочтения |


### Границы владения данными

| Сервис | База данных | Владеет данными |
|--------|-------------|-----------------|
| User Service | User DB (PostgreSQL), Session Storage (Redis) | Пользователи, профили, роли, сессии |
| Catalog Service | Catalog DB (PostgreSQL) | Товары, категории, цены |
| Order Service | Order DB (PostgreSQL) | Заказы, статусы, история |
| Payment Service | Payment DB (PostgreSQL) | Платежи, транзакции |
| Feed Service | Feed Cache (Redis) | Кэширует рекомендации в Redis |
| Notification Service | — | Не хранит данные, только маршрутизирует |

### Взаимодействие сервисов

| Тип | Механизм | Пример |
|-----|----------|--------|
| **Синхронное** | gRPC | Order Service → Catalog Service (проверка товара) |
| **Асинхронное** | Message Broker | Payment Service → Notification Service (событие оплаты) |

### Запуск

```bash
# Сборка и запуск (в src/)
docker-compose up -d --build

# Проверка статуса
docker-compose ps
```
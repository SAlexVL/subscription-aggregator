# Subscription Aggregator API

REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Стек

- Go + Gin
- PostgreSQL 16
- goose (миграции)
- swaggo (Swagger UI)
- Docker Compose

## Быстрый старт

```bash
cp .env.example .env
docker compose up --build -d
```

- API: http://localhost:8080  
- Swagger: http://localhost:8080/swagger/index.html  
- Health: http://localhost:8080/health

## API

| Метод  | Путь                         | Описание                                      |
|--------|------------------------------|-----------------------------------------------|
| POST   | `/api/v1/subscriptions`      | Создать запись                                |
| GET    | `/api/v1/subscriptions`      | Список по `user_id` (обязателен) + пагинация  |
| GET    | `/api/v1/subscriptions/{id}` | Получить по ID                                |
| PUT/PATCH | `/api/v1/subscriptions/{id}` | Частичное обновление (один SQL UPDATE)        |
| DELETE | `/api/v1/subscriptions/{id}` | Удалить                                       |
| GET    | `/api/v1/subscriptions/sum`  | Сумма за период (`from`, `to` = `MM-YYYY`)    |

### Пример создания

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

### Пример суммы

```bash
curl "http://localhost:8080/api/v1/subscriptions/sum?from=07-2025&to=09-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"
```

Сумма = `price × число месяцев пересечения` подписки с периодом `[from, to]`.
Расчёт и фильтрация выполняются в PostgreSQL (`SUM` + пересечение дат в SQL).

## Конфигурация

- `config.yaml` — значения по умолчанию
- `.env` — переопределение (`SERVER_*`, `DB_*`, `LOG_LEVEL`)

## Остановка

```bash
docker compose down
```

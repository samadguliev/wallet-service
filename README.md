# Wallet Service

Упрощённый ledger-сервис на Go — принимает финансовые транзакции (списание / зачисление) по балансу игрока, поддерживает отмену и гарантирует корректность под повторами и конкурентными запросами.

## Стек

- **Go 1.22+**
- **PostgreSQL 16** — хранилище
- **GORM** — ORM
- **Chi** — HTTP-роутер
- **shopspring/decimal** — точная арифметика для денежных операций
- **Docker Compose** — локальный запуск

## Почему PostgreSQL

Выбран PostgreSQL как наиболее подходящий для финансового сервиса:

- `NUMERIC(19,4)` — точное хранение денежных сумм без потерь точности в отличие от `FLOAT`
- `SELECT FOR UPDATE` — блокировка строк для защиты от гонок при конкурентных запросах
- `ON CONFLICT DO NOTHING` на `unique constraint` — атомарная идемпотентность на уровне БД
- транзакции с гарантией атомарности изменения баланса и записи транзакции

## Как обеспечивается корректность

**Идемпотентность** — `unique constraint` на `transactions.uuid`. При повторном запросе с тем же `transactionId` INSERT возвращает `RowsAffected=0`, сервис возвращает текущий баланс без изменений.

**Атомарность** — изменение баланса игрока и запись транзакции происходят в одной транзакции БД. Либо оба изменения применяются, либо ни одно.

**Защита от гонок** — `SELECT FOR UPDATE` на строке игрока выстраивает конкурентные запросы в очередь. Второй запрос читает уже обновлённый баланс после того как первый закоммитил.

**Порядок локов** (для предотвращения дедлоков):
```
Apply:  INSERT(transaction) → SELECT FOR UPDATE(player)
Cancel: SELECT FOR UPDATE(transaction) → SELECT FOR UPDATE(player)
```

## Структура проекта

```
wallet-service/
├── cmd/app/            # точка входа
├── internal/
│   ├── apiserver/      # HTTP-сервер, инициализация зависимостей
│   ├── domain/         # бизнес-ошибки
│   ├── handler/        # HTTP-хендлеры, middleware, response
│   ├── models/         # GORM-модели
│   ├── repository/     # работа с БД
│   └── service/        # бизнес-логика
├── .env.example
├── docker-compose.yml
└── Dockerfile
```

## Быстрый старт

**Требования:** Docker, Docker Compose

```bash
# 1. Клонировать репозиторий
git clone https://github.com/samadguliev/wallet-service.git
cd wallet-service

# 2. Создать .env из примера
cp .env.example .env

# 3. Запустить
docker compose up --build
```

Сервис будет доступен на `http://localhost:8080`.

## Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|---|---|---|
| `APP_PORT` | Порт HTTP-сервера | `8080` |
| `DB_HOST` | Хост PostgreSQL | `postgres` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_NAME` | Имя базы данных | `wallet` |
| `DB_USER` | Пользователь БД | `wallet` |
| `DB_PASSWORD` | Пароль БД | `wallet123` |
| `AUTH_TOKEN` | Bearer-токен для авторизации | `super-secret-token` |

## API

Все эндпоинты требуют заголовок авторизации:
```
Authorization: Bearer <AUTH_TOKEN>
```

### GET /players/{id}/balance

Получить баланс игрока.

```bash
curl http://localhost:8080/players/550e8400-e29b-41d4-a716-446655440000/balance \
  -H "Authorization: Bearer super-secret-token"
```

```json
{ "data": { "balance": "100.00" } }
```

### POST /transactions

Провести транзакцию (списание или зачисление).

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Authorization: Bearer super-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "transactionId": "7f8d9e0a-1b2c-3d4e-5f6a-7b8c9d0e1f2a",
    "playerId": "550e8400-e29b-41d4-a716-446655440000",
    "type": "withdraw",
    "amount": "10.50",
    "currency": "eur"
  }'
```

```json
{ "data": { "balance": "89.50" } }
```

| Поле | Тип | Описание |
|---|---|---|
| `transactionId` | uuid | Ключ идемпотентности — повтор запроса не применяет операцию повторно |
| `playerId` | uuid | ID игрока |
| `type` | string | `withdraw` или `deposit` |
| `amount` | string | Сумма (не отрицательная, `0` допустим) |
| `currency` | string | Валюта игрока (должна совпадать) |

### DELETE /transactions/{transactionId}

Отменить транзакцию. Работает только для `withdraw`. Идемпотентна — повторная отмена возвращает текущий баланс.

```bash
curl -X DELETE http://localhost:8080/transactions/7f8d9e0a-1b2c-3d4e-5f6a-7b8c9d0e1f2a \
  -H "Authorization: Bearer super-secret-token"
```

```json
{ "data": { "balance": "100.00" } }
```

## Формат ошибок

```json
{
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "insufficient funds for withdrawal"
  }
}
```

| Код | HTTP | Описание |
|---|---|---|
| `UNAUTHORIZED` | 401 | Отсутствует или невалидный токен |
| `VALIDATION_ERROR` | 400 | Некорректные входные данные |
| `INSUFFICIENT_FUNDS` | 422 | Недостаточно средств для списания |
| `DEPOSIT_CANNOT_CANCEL` | 422 | Попытка отменить deposit |
| `NOT_FOUND` | 404 | Игрок или транзакция не найдены |
| `TRANSACTION_FAILED` | 500 | Внутренняя ошибка сервера |

## Тесты

```bash
# Все тесты
go test ./...

# С выводом
go test ./... -v

# С покрытием
go test ./... -cover
```

Покрыты: идемпотентность Apply и Cancel, недостаток средств, отмена deposit, повторная отмена, несовпадение валюты, граничные случаи (нулевой баланс, нулевая сумма).

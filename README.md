# Personal Finance Analytics

Учебный production-like проект на Go для учёта личных расходов и анализа трат.

Проект состоит из Gateway, Ledger и Auth, которые взаимодействуют по gRPC. Внешний API Gateway — HTTP.

## Архитектура

```text
HTTP client
    |
    v
gateway :8080
    |-------------- gRPC ------------> auth :9091 ----> PostgreSQL
    |
    |-------------- gRPC ------------> ledger :9090 --> PostgreSQL
                                                  \
                                                   --> Redis
```

Модули workspace:

- `gateway` — HTTP API, JWT middleware, Auth/Ledger gRPC clients;
- `ledger` — budgets, transactions, reports, CSV import/export и Redis cache;
- `auth` — регистрация, login, bcrypt и JWT issue;
- `shared` — protobuf contracts и generated Go code.

Бизнес-слой не зависит от PostgreSQL, Redis, HTTP или gRPC. Persistence разделён на database abstraction, repositories и database adapters.

## Требования

- Go 1.26+;
- Docker Compose или Podman Compose;
- Task опционален;
- PostgreSQL и Redis локально устанавливать не нужно.

## Конфигурация

```bash
cp .env.example .env
```

Основные значения по умолчанию:

```dotenv
POSTGRES_DB=finance
POSTGRES_USER=finance
POSTGRES_PASSWORD=finance
POSTGRES_PORT=5432

LEDGER_DATABASE_URL=postgres://finance:finance@localhost:5432/finance?sslmode=disable
AUTH_DATABASE_URL=postgres://finance:finance@localhost:5432/finance?sslmode=disable

LEDGER_GRPC_PORT=9090
AUTH_GRPC_PORT=9091
REDIS_HOST=localhost
REDIS_PORT=6379
```

`AUTH_JWT_SECRET` должен содержать минимум 32 байта. Не используйте production secret из `.env.example`.

## Запуск через Compose

```bash
cp .env.example .env
docker compose up -d --build
```

Для Podman:

```bash
podman compose up -d --build
```

Проверка Gateway:

```bash
curl http://localhost:8080/ping
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

После первого запуска примените миграции:

```bash
task migration:up
task migration:auth:up
```

Без Task эквивалентные команды выполняются локальным бинарником Goose:

```bash
./bin/goose -dir migrations/ledger postgres "$LEDGER_DATABASE_URL" up
./bin/goose -table goose_auth_db_version -dir migrations/auth postgres "$AUTH_DATABASE_URL" up
```

Остановка:

```bash
docker compose down
# или
podman compose down
```

## OpenAPI

Спецификация доступна после запуска Gateway:

```text
http://localhost:8080/openapi.yaml
```

Файл можно импортировать в Postman или другой API client:

```bash
curl http://localhost:8080/openapi.yaml -o openapi.yaml
```

## JWT flow

Зарегистрируйте пользователя:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secure-password"}'
```

Получите access token:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secure-password"}'
```

Ответ login содержит access и refresh токены. Если access token истёк, получите новый без повторного ввода пароля:

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refresh_token>"}'
```

Передавайте токен в защищённые endpoints:

```bash
export ACCESS_TOKEN='<access_token>'
curl http://localhost:8080/budgets \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

`user_id` для Ledger-запросов не принимается от клиента. Gateway получает identity из JWT и передаёт её в Ledger.

## HTTP API

### Budgets

Создать или обновить бюджет:

```bash
curl -X PUT http://localhost:8080/budgets/food \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"limit_amount":3000}'
```

Получить бюджеты текущего пользователя:

```bash
curl http://localhost:8080/budgets \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Transactions

Создать транзакцию. Денежные суммы передаются в минимальных денежных единицах и хранятся как `int64`:

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "amount": 1500,
    "category": "food",
    "description": "lunch",
    "occurred_at": "2026-10-01T12:00:00Z"
  }'
```

Получить транзакции за период:

```bash
curl 'http://localhost:8080/transactions?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z&category=food' \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Reports

```bash
curl 'http://localhost:8080/reports/summary?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z' \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Отчёты кешируются в Redis с TTL. В логах Ledger отображаются cache hit/miss.

### CSV

CSV format:

```csv
amount,category,description,occurred_at
1500,food,lunch,2026-10-01T12:00:00Z
```

Импорт:

```bash
curl -X POST http://localhost:8080/transactions/import \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: text/csv' \
  --data-binary @transactions.csv
```

Экспорт:

```bash
curl 'http://localhost:8080/transactions/export?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z' \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -o exported-transactions.csv
```

`user_id` отсутствует в CSV и определяется по JWT. Максимальный размер HTTP import — 10 MiB.

## Миграции

Создание Ledger migration без подключения к БД:

```bash
task migration:create NAME=create_example
```

Статус и применение:

```bash
task migration:status
task migration:up
task migration:down

task migration:auth:status
task migration:auth:up
task migration:auth:down
```

Используется Goose `v3.28.0`.

## Проверки

Все developer tools устанавливаются в `./bin`:

```bash
task tools
task format
task test
task test:race
task test:integration
task lint
task proto:lint
task proto:gen
task mockery:gen
```

Integration tests требуют работающие PostgreSQL и Redis и запускаются отдельно от unit tests.

Эквивалент unit/race проверки без Task:

```bash
go test ./...
go test -race ./...
```

## Полный demo-сценарий

```bash
# 1. Запуск инфраструктуры и сервисов
cp .env.example .env
docker compose up -d --build
task migration:up
task migration:auth:up

# 2. Регистрация и login
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secure-password"}'

LOGIN_RESPONSE=$(curl --silent -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"secure-password"}')
export ACCESS_TOKEN=$(printf '%s' "$LOGIN_RESPONSE" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

# 3. Бюджет и транзакция
curl -X PUT http://localhost:8080/budgets/food \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"limit_amount":3000}'

curl -X POST http://localhost:8080/transactions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount":1500,"category":"food","description":"lunch","occurred_at":"2026-10-01T12:00:00Z"}'

# 4. Отчёт и CSV
curl "http://localhost:8080/reports/summary?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl "http://localhost:8080/transactions/export?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -o exported-transactions.csv
```

Для shell-скриптов, где email уже существует, пропустите регистрацию и выполните только login.

## Пошаговая проверка сервисов

### 1. Подготовить окружение

```bash
cp .env.example .env
docker compose up -d --build
# или: podman compose up -d --build
```

Проверьте, что все контейнеры запущены:

```bash
docker compose ps
```

PostgreSQL и Redis должны иметь статус `healthy`, Auth и Ledger должны быть запущены, Gateway должен слушать порт `8080`.

### 2. Применить миграции

```bash
task migration:up
task migration:auth:up
task migration:status
task migration:auth:status
```

В статусе должны присутствовать две Ledger migration и одна Auth migration.

### 3. Проверить Gateway и OpenAPI

```bash
curl --fail http://localhost:8080/ping
curl --fail http://localhost:8080/openapi.yaml | head
```

Ожидается `{"status":"ok"}` и документ, начинающийся с `openapi: 3.0.3`.

### 4. Проверить Auth

```bash
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"check@example.com","password":"secure-password"}'

curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"check@example.com","password":"secure-password"}'
```

Сохраните `access_token` из ответа login:

```bash
export ACCESS_TOKEN='<access_token>'
```

Повторная регистрация того же email должна вернуть `409`, а login с неверным паролем — `401`.

### 5. Проверить Ledger ownership и budgets

```bash
curl -X PUT http://localhost:8080/budgets/food \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"limit_amount":3000}'

curl --fail http://localhost:8080/budgets \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

В JSON-запросе нельзя передавать `user_id`: identity берётся из JWT.

### 6. Проверить transactions и budget invariant

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"amount":1500,"category":"food","description":"check","occurred_at":"2026-10-01T12:00:00Z"}'

curl 'http://localhost:8080/transactions?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z' \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Транзакция, превышающая остаток бюджета, должна завершиться `409 Conflict`.

### 7. Проверить reports и Redis cache

```bash
curl 'http://localhost:8080/reports/summary?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z' \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl 'http://localhost:8080/reports/summary?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z' \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

В логах Ledger первый запрос должен показать `summary cache miss`, а повторный — `summary cache hit`.

### 8. Проверить CSV import/export

Создайте файл:

```bash
cat > transactions.csv <<'CSV'
amount,category,description,occurred_at
500,food,coffee,2026-10-02T10:00:00Z
CSV
```

Импортируйте его:

```bash
curl -X POST http://localhost:8080/transactions/import \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: text/csv' \
  --data-binary @transactions.csv
```

Ожидаемый ответ содержит `imported_count: 1`.

Экспортируйте период:

```bash
curl 'http://localhost:8080/transactions/export?from=2026-10-01T00:00:00Z&to=2026-11-01T00:00:00Z' \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -o exported-transactions.csv
```

В exported CSV должны присутствовать исходные заголовки и импортированная транзакция.

### 9. Запустить автоматические проверки

```bash
task test
task test:race
task test:integration
task lint
```

`task test:integration` проверяет PostgreSQL, Redis, Auth gRPC, Ledger gRPC и полный Gateway HTTP сценарий.

### 10. Проверить Google Sheets

1. Откройте Google Sheet → `Extensions` → `Apps Script`.
2. Скопируйте `integrations/google-sheets/Code.gs`.
3. Выполните `setGatewayConfig` с доступным из Google Apps Script Gateway URL и JWT.
4. Создайте строку с колонками Date, Amount, Category, Description.
5. Запустите `createTransactionFromActiveRow`.
6. Запустите `writeCurrentMonthSummary` и проверьте колонки H:I.

## Google Sheets

Apps Script находится в [integrations/google-sheets/Code.gs](integrations/google-sheets/Code.gs).

Структура листа для создания транзакции из активной строки:

| A: Date | B: Amount | C: Category | D: Description | F: Result |
|---|---:|---|---|---|
| `2026-10-01T12:00:00Z` | `1500` | `food` | `lunch` | transaction ID |

Настройка:

1. Откройте Google Sheet и выберите `Extensions → Apps Script`.
2. Скопируйте содержимое `integrations/google-sheets/Code.gs`.
3. Получите JWT через `/auth/login`.
4. Один раз выполните в редакторе Apps Script:

```javascript
setGatewayConfig('https://your-gateway.example.com', '<jwt-access-token>');
```

Для локального Gateway используйте адрес, доступный из браузера/Apps Script. `localhost` из Apps Script указывает на инфраструктуру Google, а не на ваш компьютер.

Доступные функции:

- `createTransaction(amount, category, description, occurredAt)` — создаёт транзакцию;
- `createTransactionFromActiveRow()` — берёт значения A:D активной строки;
- `getSummary(from, to)` — получает отчёт за период;
- `writeCurrentMonthSummary()` — записывает текущий отчёт в колонки H:I.

JWT хранится в `UserProperties` и не записывается в таблицу. После истечения access token повторите login и выполните `setGatewayConfig` с новым токеном.

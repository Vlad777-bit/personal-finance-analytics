# Personal Finance Analytics

Учебный микросервисный проект для учёта и анализа личных трат.

## Текущий этап

Репозиторий организован как Go workspace с отдельными модулями `gateway`, `ledger`, `auth` и `shared`.
На текущей итерации рабочим HTTP-процессом является Gateway с `GET /ping`; PostgreSQL и Redis запускаются через Compose. Ledger и Auth пока содержат минимальные entrypoint-заглушки и будут развиваться отдельными итерациями.

## Требования

- Go 1.26+ (проект фиксирует минимальную версию Go 1.26; разработка на Go 1.27 также допустима)
- Docker Compose-совместимый CLI или Podman с Compose
- Task — опционально, используется только как удобный developer-интерфейс

## Структура workspace

```text
.
├── auth/
├── gateway/
├── ledger/
├── shared/
├── bin/
├── go.work
├── compose.yaml
├── Dockerfile
└── Taskfile.yml
```

Go-модули:

- `github.com/Vlad777-bit/personal-finance-analytics/gateway`
- `github.com/Vlad777-bit/personal-finance-analytics/ledger`
- `github.com/Vlad777-bit/personal-finance-analytics/auth`
- `github.com/Vlad777-bit/personal-finance-analytics/shared`

## Быстрый старт без Task

```bash
cp .env.example .env
docker compose -f compose.yaml up -d --build
curl http://localhost:8080/ping
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

Остановка:

```bash
docker compose -f compose.yaml down
```

## Проверки без Task

Из корня репозитория:

```bash
go work sync

for module in gateway ledger auth shared; do
  (cd "$module" && go mod tidy)
done

for module in gateway ledger auth shared; do
  if (cd "$module" && go list ./... 2>/dev/null | grep -q .); then
    (cd "$module" && go test ./...)
  fi
done

for module in gateway ledger auth shared; do
  if (cd "$module" && go list ./... 2>/dev/null | grep -q .); then
    (cd "$module" && go test -race ./...)
  fi
done
```

## Через Task

```bash
task tools
task deps:update
task format
task test
task test:race
task lint
task up
```

Все внешние developer-инструменты устанавливаются в `./bin`. Содержимое `bin` не коммитится, кроме `.gitkeep`.

## Podman

Если `docker` указывает на Docker-compatible Podman CLI, команды из README работают без изменений. Иначе используй эквивалент:

```bash
podman compose -f compose.yaml up -d --build
podman compose -f compose.yaml down
```

## Protobuf

Buf-конфигурация находится в `shared/proto`.

```bash
task proto:lint
task proto:gen
```

Generated Go-код будет помещаться в `shared/gen/go`.

## Mockery

Mockery устанавливается локально в `./bin`. На текущем этапе `packages` в `.mockery.yaml` пуст, потому что интерфейсы Ledger/Auth ещё не введены. Они будут добавлены вместе с соответствующей бизнес-логикой, чтобы конфигурация не ссылалась на несуществующие пакеты.

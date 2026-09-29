# Personal Finance Analytics

Учебный микросервисный проект для учёта и анализа личных трат.

## Требования

- Go 1.26+
- Docker Compose-совместимый CLI (Docker Compose или Podman с Docker-compatible command)
- Task — опционально, только для удобства разработки

## Быстрый старт без Task

```bash
cp .env.example .env
docker compose up -d --build
curl http://localhost:8080/ping
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

Тесты можно запускать отдельно:

```bash
go test ./...
go test -race ./...
```

## Через Task

```bash
task tools
task test
task up
```

Все внешние dev-инструменты устанавливаются в `./bin` и не загрязняют систему.

## Podman

Compose-файл не использует Docker-specific API. Если в системе Podman предоставляет Docker-compatible CLI, команды выше работают без изменений. В ином случае можно использовать эквивалентную команду `podman compose`.

## Статус архитектуры

Первая итерация содержит каркас проекта, контейнеризованный Gateway с `GET /ping`, PostgreSQL и Redis в Compose, а также локальный toolchain. Ledger/Auth и прикладные API будут добавляться следующими итерациями.

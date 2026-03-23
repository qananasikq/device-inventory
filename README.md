# device-inventory

Учёт сетевых устройств. Go + SQLite, React + TypeScript.


## Запуск

```
docker compose up --build
```

Фронт — http://localhost:5173, бэкенд — http://localhost:8080.

База создаётся автоматически в `data/`.

## Зависимости

Backend: Go 1.24+ (CGO для sqlite3), Frontend: Node 18+.

Без докера:

```
cd backend
go run ./cmd/main.go
```

```
cd frontend
npm i && npm run dev
```

По умолчанию dev-сервер frontend проксирует API на `http://localhost:8080`.
При необходимости можно переопределить адрес через переменную `VITE_API_URL`.

Мок-режим (без бэкенда, данные в памяти):

```
cd frontend
VITE_MOCK_API=true npm run dev
```

## API

- `GET /health` — healthcheck
- `GET /devices` — список, фильтры `is_active` и `search`
- `GET /devices/:id` — одно устройство
- `POST /devices` — создать
- `PUT /devices/:id` — обновить
- `DELETE /devices/:id` — soft delete (деактивация)

### Примеры

Создать устройство:
```
curl -X POST http://localhost:8080/devices \
  -H "Content-Type: application/json" \
  -d '{"hostname":"gw-msk-01","ip":"192.168.1.10","location":"Moscow","is_active":true}'
```

Список активных с поиском:
```
curl "http://localhost:8080/devices?is_active=true&search=gw"
```

Обновить:
```
curl -X PUT http://localhost:8080/devices/1 \
  -H "Content-Type: application/json" \
  -d '{"hostname":"gw-msk-01","ip":"192.168.1.11","location":"SPB","is_active":true}'
```

Деактивировать:
```
curl -X DELETE http://localhost:8080/devices/1
```

## Тесты

```
cd backend && go test ./...
```

## SQL

В папке `sql/`:

- `schema.sql` — схема БД (devices, configs, logs)
- `queries.sql` — запросы, индексы и оптимизация

## Env-переменные

- `DB_CONNECTION` — путь к sqlite (по умолчанию `device_inventory.db`)
- `PORT` — порт бэкенда (по умолчанию `8080`)
- `CORS_ORIGIN` — allowed origin (по умолчанию `*`)
- `VITE_MOCK_API` — `true` чтобы фронт работал без бэкенда

## Задание 4 — исправленные фрагменты

В папке `task/` лежат исправленные фрагменты кода из задания 4:

- `task4-fragment-go-wave-1-fixed.go` — Go-фрагмент (контекст, обработка ошибок, аудит)
- `task4-fragment-ts-wave-1-fixed.ts` — TypeScript-фрагмент (fetch, debounce, AbortController)



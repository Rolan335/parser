# parser

Сервис принимает файл, снимает метаданные через утилиту exiftool, кладёт в Postgres и возвращает клиенту JSON с полями метаданных и тот же json в виде файла в base64.

Стек: Go 1.25, gin, pgx/v5, barasher/go-exiftool, caarlos0/env, log/slog, Postgres 16.

## Запуск

```bash
docker compose up --build
```

В compose env подхватывается из `.env.example` через `env_file`. 

Env:
- `HTTP_ADDR` — адрес HTTP-сервера, по умолчанию `:8080`
- `DATABASE_URL` — обязательный

## API

### POST /suip-data/parse

multipart-форма, поле `file`. Лимит запроса 100 MiB.

```bash
curl -F file=@./test.pdf http://localhost:8080/suip-data/parse
```

Коды: `200` ok, `400` нет файла / слишком большой / битые multipart-данные, `422` exiftool не справился, `500` всё остальное.

Ответ:
```json
{
  "file_name": "test.pdf", "size_bytes": 93210,
  "mime_type": "application/pdf", "format": "PDF",
  "title": "...", "producer": "...",
  "raw": {...}, "created_at": "2026-05-27T12:34:56Z",
  "file": { "name": "test.pdf.meta.json", "content_base64": "..." }
}
```

Достать файл:
```bash
curl -s -F file=@./test.pdf http://localhost:8080/suip-data/parse \
  | jq -r '.file.content_base64' | base64 -d > meta.json
```

### GET /suip-data

Query-параметры (все опциональные):
- `file_name`, `mime_type`, `format` — точное совпадение
- `from`, `to` — RFC3339, диапазон по `created_at`
- `limit`, `offset` — пагинация, `limit=0` означает без лимита

```bash
curl 'http://localhost:8080/suip-data?format=PDF&limit=10'
```

Ответ: `{ "items": [...], "count": N }`.

## Структура

```
cmd/server/          точка входа
internal/config/     env-структура
internal/domain/     модели
internal/extractor/  обёртка над exiftool
internal/repository/ pgxpool, SQL в const.go
internal/service/    бизнес-логика, типизированные ошибки
internal/handler/    gin-хендлеры
migrations/          001_init.sql
```

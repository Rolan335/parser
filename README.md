# parser

Сервис принимает PDF, прогоняет его через https://www.pdfyeah.com/view-pdf-metadata, вытаскивает блок `<textarea>` с результатом, парсит его в поля, кладёт в Postgres и возвращает клиенту JSON с этими полями плюс сам textarea-сниппет в base64.

Стек: Go 1.25, gin, pgx/v5, caarlos0/env, log/slog, Postgres 16.

## Запуск

```bash
docker compose up --build
```

В compose env подхватывается из `.env.example` через `env_file`.

Env:
- `HTTP_ADDR` — адрес HTTP-сервера, по умолчанию `:8080`
- `DATABASE_URL` — обязательный

Если меняется схема — снести volume:
```bash
docker compose down -v && docker compose up --build
```

## API

### POST /pdfyear-data/parse

multipart-форма, поле `file`. Лимит запроса 100 MiB.

```bash
curl -F file=@./test.pdf http://localhost:8080/pdfyear-data/parse
```

Коды: `200` ok, `400` нет файла / слишком большой / битые multipart-данные, `422` pdfyeah не отдал блок метаданных, `500` всё остальное.

Ответ:
```json
{
  "id": 7,
  "file_name": "test.pdf",
  "size_bytes": 35926,
  "producer": "cairo 1.16.0 (https://cairographics.org)",
  "title": "Hello",
  "creation_date": "Mon Feb 16 10:14:18 2026 UTC",
  "pages": 2,
  "pdf_version": "1.5",
  "page_size": "595.276 x 841.89 pts (A4)",
  "page_rot": 0,
  "form": "none",
  "encrypted": false,
  "optimized": false,
  "tagged": false,
  "javascript": false,
  "custom_metadata": false,
  "metadata_stream": false,
  "user_properties": false,
  "suspects": false,
  "created_at": "2026-05-30T14:21:37Z",
  "file": {
    "name": "test.pdf.html",
    "content_base64": "PHRleHRhcmVhIC4uLg=="
  }
}
```

`size_bytes` берётся из строки `File size` в ответе pdfyeah; если её нет — подставляется размер загруженного файла. `file.content_base64` — это исходный HTML-сниппет с `<textarea>...</textarea>` ровно как его отдал pdfyeah.

Достать сниппет в файл:
```bash
curl -s -F file=@./test.pdf http://localhost:8080/pdfyear-data/parse \
  | jq -r '.file.content_base64' | base64 -d > meta.html
```

### GET /pdfyear-data

Query-параметры (все опциональные):
- `file_name`, `producer`, `pdf_version` — точное совпадение
- `from`, `to` — RFC3339, диапазон по `created_at`
- `limit`, `offset` — пагинация, `limit=0` означает без лимита

```bash
curl 'http://localhost:8080/pdfyear-data?producer=cairo%201.16.0&limit=10'
```

Ответ: `{ "items": [...], "count": N }`. Сниппет (`raw_html`) в БД хранится, но в списке не отдаётся.

## Структура

```
cmd/server/          точка входа
internal/config/     env-структура
internal/domain/     модели
internal/extractor/  HTTP-клиент к pdfyeah + парсер HTML
internal/repository/ pgxpool, SQL в const.go
internal/service/    бизнес-логика, типизированные ошибки
internal/handler/    gin-хендлеры
migrations/          001_init.sql
```

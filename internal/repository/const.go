package repository

const insertFileMetadataSQL = `
INSERT INTO file_metadata
	(file_name, size_bytes, mime_type, format, title, producer, raw, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id`

const listFileMetadataSQL = `
SELECT id, file_name, size_bytes, mime_type, format, title, producer, raw, created_at
FROM file_metadata
WHERE ($1 = '' OR file_name = $1)
  AND ($2 = '' OR mime_type = $2)
  AND ($3 = '' OR format = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5)
ORDER BY id DESC
LIMIT NULLIF($6, 0)
OFFSET $7`

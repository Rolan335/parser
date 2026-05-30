package repository

const insertFileMetadataSQL = `
INSERT INTO file_metadata
	(file_name, size_bytes, producer, title, creation_date,
	 pages, pdf_version, page_size, page_rot, form,
	 encrypted, optimized, tagged, javascript,
	 custom_metadata, metadata_stream, user_properties, suspects,
	 raw_html, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
        $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
RETURNING id`

const listFileMetadataSQL = `
SELECT id, file_name, size_bytes, producer, title, creation_date,
       pages, pdf_version, page_size, page_rot, form,
       encrypted, optimized, tagged, javascript,
       custom_metadata, metadata_stream, user_properties, suspects,
       raw_html, created_at
FROM file_metadata
WHERE ($1 = '' OR file_name = $1)
  AND ($2 = '' OR producer = $2)
  AND ($3 = '' OR pdf_version = $3)
  AND ($4::timestamptz IS NULL OR created_at >= $4)
  AND ($5::timestamptz IS NULL OR created_at <= $5)
ORDER BY id DESC
LIMIT NULLIF($6, 0)
OFFSET $7`

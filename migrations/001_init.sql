CREATE TABLE IF NOT EXISTS file_metadata (
    id          BIGSERIAL    PRIMARY KEY,
    file_name   VARCHAR(255) NOT NULL,
    size_bytes  BIGINT       NOT NULL,
    mime_type   VARCHAR(127) NOT NULL DEFAULT '',
    format      VARCHAR(32)  NOT NULL DEFAULT '',
    title       VARCHAR(512) NOT NULL DEFAULT '',
    producer    VARCHAR(255) NOT NULL DEFAULT '',
    raw         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS file_metadata_created_at_idx ON file_metadata (created_at DESC);
CREATE INDEX IF NOT EXISTS file_metadata_mime_type_idx  ON file_metadata (mime_type);
CREATE INDEX IF NOT EXISTS file_metadata_format_idx     ON file_metadata (format);

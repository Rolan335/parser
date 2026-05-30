CREATE TABLE IF NOT EXISTS file_metadata (
    id              BIGSERIAL    PRIMARY KEY,
    file_name       VARCHAR(255) NOT NULL,
    size_bytes      BIGINT       NOT NULL,
    producer        VARCHAR(255) NOT NULL DEFAULT '',
    title           VARCHAR(512) NOT NULL DEFAULT '',
    creation_date   VARCHAR(64)  NOT NULL DEFAULT '',
    pages           INTEGER      NOT NULL DEFAULT 0,
    pdf_version     VARCHAR(16)  NOT NULL DEFAULT '',
    page_size       VARCHAR(64)  NOT NULL DEFAULT '',
    page_rot        INTEGER      NOT NULL DEFAULT 0,
    form            VARCHAR(32)  NOT NULL DEFAULT '',
    encrypted       BOOLEAN      NOT NULL DEFAULT FALSE,
    optimized       BOOLEAN      NOT NULL DEFAULT FALSE,
    tagged          BOOLEAN      NOT NULL DEFAULT FALSE,
    javascript      BOOLEAN      NOT NULL DEFAULT FALSE,
    custom_metadata BOOLEAN      NOT NULL DEFAULT FALSE,
    metadata_stream BOOLEAN      NOT NULL DEFAULT FALSE,
    user_properties BOOLEAN      NOT NULL DEFAULT FALSE,
    suspects        BOOLEAN      NOT NULL DEFAULT FALSE,
    raw_html        BYTEA        NOT NULL DEFAULT ''::bytea,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS file_metadata_created_at_idx  ON file_metadata (created_at DESC);
CREATE INDEX IF NOT EXISTS file_metadata_producer_idx    ON file_metadata (producer);
CREATE INDEX IF NOT EXISTS file_metadata_pdf_version_idx ON file_metadata (pdf_version);

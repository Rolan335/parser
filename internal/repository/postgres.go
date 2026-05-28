package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rolan335/parser/internal/domain"
)

type Repository interface {
	Save(ctx context.Context, m *domain.FileMetadata) (int64, error)
	List(ctx context.Context, f domain.ListFilter) ([]domain.FileMetadata, error)
}

type PgRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

func (r *PgRepository) Save(ctx context.Context, m *domain.FileMetadata) (int64, error) {
	raw, err := json.Marshal(m.Raw)
	if err != nil {
		return 0, fmt.Errorf("marshal raw: %w", err)
	}
	var id int64
	err = r.pool.QueryRow(ctx, insertFileMetadataSQL,
		m.FileName, m.SizeBytes, m.MimeType, m.Format,
		m.Title, m.Producer, raw, m.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert metadata: %w", err)
	}
	return id, nil
}

func (r *PgRepository) List(ctx context.Context, f domain.ListFilter) ([]domain.FileMetadata, error) {
	rows, err := r.pool.Query(ctx, listFileMetadataSQL,
		f.FileName, f.MimeType, f.Format,
		f.From, f.To,
		f.Limit, f.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query metadata: %w", err)
	}
	defer rows.Close()

	out := make([]domain.FileMetadata, 0)
	for rows.Next() {
		var m domain.FileMetadata
		var raw []byte
		if err := rows.Scan(&m.ID, &m.FileName, &m.SizeBytes, &m.MimeType, &m.Format,
			&m.Title, &m.Producer, &raw, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &m.Raw); err != nil {
				return nil, fmt.Errorf("unmarshal raw id=%d: %w", m.ID, err)
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

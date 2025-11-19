package dns

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Record struct {
	ID        int       `json:"id"`
	DomainID  int       `json:"domain_id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	TTL       int       `json:"ttl"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	GetRecords(ctx context.Context, domainID int) ([]*Record, error)
	GetRecordsByNameAndType(ctx context.Context, name, recordType string) ([]*Record, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetRecords(ctx context.Context, domainID int) ([]*Record, error) {
	query := `SELECT id, domain_id, type, name, content, ttl, priority, created_at FROM dns_records WHERE domain_id = $1`
	rows, err := r.db.Query(ctx, query, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		rec := &Record{}
		if err := rows.Scan(&rec.ID, &rec.DomainID, &rec.Type, &rec.Name, &rec.Content, &rec.TTL, &rec.Priority, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *repository) GetRecordsByNameAndType(ctx context.Context, name, recordType string) ([]*Record, error) {
	// This is a simplified query. Real DNS logic needs to handle wildcards, etc.
	// Also, we need to join with domains table to ensure the domain is active and exists.
	query := `
		SELECT r.id, r.domain_id, r.type, r.name, r.content, r.ttl, r.priority, r.created_at 
		FROM dns_records r
		JOIN domains d ON r.domain_id = d.id
		WHERE r.name = $1 AND r.type = $2 AND d.active = TRUE
	`
	rows, err := r.db.Query(ctx, query, name, recordType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		rec := &Record{}
		if err := rows.Scan(&rec.ID, &rec.DomainID, &rec.Type, &rec.Name, &rec.Content, &rec.TTL, &rec.Priority, &rec.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

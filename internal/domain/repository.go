package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Domain struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	TargetURL string          `json:"target_url"`
	Active    bool            `json:"active"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, d *Domain) error
	GetByName(ctx context.Context, name string) (*Domain, error)
	GetAll(ctx context.Context) ([]*Domain, error)
	Delete(ctx context.Context, id int) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, d *Domain) error {
	query := `
		INSERT INTO domains (name, target_url, active, config_json)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, d.Name, d.TargetURL, d.Active, d.Config).
		Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (r *repository) GetByName(ctx context.Context, name string) (*Domain, error) {
	query := `SELECT id, name, target_url, active, config_json, created_at, updated_at FROM domains WHERE name = $1`
	d := &Domain{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&d.ID, &d.Name, &d.TargetURL, &d.Active, &d.Config, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*Domain, error) {
	query := `SELECT id, name, target_url, active, config_json, created_at, updated_at FROM domains`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*Domain
	for rows.Next() {
		d := &Domain{}
		if err := rows.Scan(&d.ID, &d.Name, &d.TargetURL, &d.Active, &d.Config, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM domains WHERE id = $1", id)
	return err
}

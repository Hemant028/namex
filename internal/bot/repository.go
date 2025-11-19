package bot

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BotRule struct {
	ID          int       `json:"id"`
	RuleType    string    `json:"rule_type"`
	Value       string    `json:"value"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, rule *BotRule) error
	GetAll(ctx context.Context) ([]*BotRule, error)
	Delete(ctx context.Context, id int) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, rule *BotRule) error {
	query := `
		INSERT INTO bot_rules (rule_type, value, action, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, rule.RuleType, rule.Value, rule.Action, rule.Description).
		Scan(&rule.ID, &rule.CreatedAt)
}

func (r *repository) GetAll(ctx context.Context) ([]*BotRule, error) {
	query := `SELECT id, rule_type, value, action, description, created_at FROM bot_rules`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*BotRule
	for rows.Next() {
		rule := &BotRule{}
		if err := rows.Scan(&rule.ID, &rule.RuleType, &rule.Value, &rule.Action, &rule.Description, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM bot_rules WHERE id = $1", id)
	return err
}

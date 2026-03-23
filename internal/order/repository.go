package order

import (
	"context"
	"database/sql"

	"go-otel-2/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, order *models.Order) error {
	query := `INSERT INTO orders (id, customer_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.CustomerID, order.Amount, order.Status, order.CreatedAt)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	query := `SELECT id, customer_id, amount, status, created_at
		FROM orders WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var o models.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.Amount, &o.Status, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

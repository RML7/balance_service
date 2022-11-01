package repo

import (
	"context"
	"errors"

	"github.com/avito-test/internal/storage/db"
	"github.com/jackc/pgx/v4"
)

type BalanceRepo struct {
	dbClient db.Client
}

func NewBalanceRepo(dbClient db.Client) BalanceRepo {
	return BalanceRepo{dbClient: dbClient}
}

func (r *BalanceRepo) AddBalance(userId string, sum float64, comment *string) error {
	sql := "SELECT public.\"add_balance\"($1, $2, $3)"

	_, err := r.dbClient.Exec(context.TODO(), sql, userId, sum, comment)
	if err != nil {
		return err
	}

	return nil
}

func (r *BalanceRepo) GetBalanceByUserId(userId string) (*float64, error) {
	sql := "SELECT balance FROM public.balance WHERE user_id = $1"

	var balance *float64

	if err := r.dbClient.QueryRow(context.TODO(), sql, userId).Scan(&balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return balance, nil
}

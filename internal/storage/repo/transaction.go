package repo

import (
	"context"

	"github.com/avito-test/internal/storage/db"
)

type TransactionRepo struct {
	dbClient db.Client
}

func NewTransactionRepo(dbClient db.Client) TransactionRepo {
	return TransactionRepo{dbClient: dbClient}
}

func (t *TransactionRepo) SaveTransaction(orderId string, userId string, serviceId string, sum float64, transactionType int, comment *string) (int, error) {
	sql := "SELECT  public.\"save_transaction\"($1,$2,$3,$4,$5::smallint,$6) as \"status\""

	var status int

	if err := t.dbClient.QueryRow(
		context.TODO(),
		sql,
		orderId,
		userId,
		serviceId,
		sum,
		transactionType,
		comment).Scan(&status); err != nil {

		return 0, err
	}

	return status, nil
}

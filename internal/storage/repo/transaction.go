package repo

import (
	"context"

	"github.com/avito-test/internal/model"
	"github.com/avito-test/internal/storage/db"
)

type TransactionRepo struct {
	dbClient db.Client
}

func NewTransactionRepo(dbClient db.Client) TransactionRepo {
	return TransactionRepo{dbClient: dbClient}
}

func (t *TransactionRepo) SaveTransaction(transaction model.Transaction) (int, error) {
	sql := "SELECT  public.\"save_transaction\"($1,$2,$3,$4,$5::smallint) as \"status\""

	var status int

	if err := t.dbClient.QueryRow(
		context.TODO(),
		sql,
		transaction.OrderId,
		transaction.UserId,
		transaction.ServiceId,
		transaction.Sum,
		transaction.TransactionTypeId).Scan(&status); err != nil {

		return 0, err
	}

	return status, nil
}

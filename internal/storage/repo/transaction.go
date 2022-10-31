package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/avito-test/internal/model"
	"github.com/avito-test/internal/storage/db"
)

type TransactionRepo struct {
	dbClient db.Client
}

func NewTransactionRepo(dbClient db.Client) TransactionRepo {
	return TransactionRepo{dbClient: dbClient}
}

func (t *TransactionRepo) SaveTransaction(orderId string, userId string, serviceId string, sum float64, transactionType int, comment *string) (int, error) {
	sqlRow := "SELECT  public.\"save_transaction\"($1,$2,$3,$4,$5::smallint,$6) as \"status\""

	var status int

	if err := t.dbClient.QueryRow(
		context.TODO(),
		sqlRow,
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

func (t *TransactionRepo) GetTransactionListByUserId(userId string, offset int, limit int, sortBy string, sortType string) ([]model.Transaction, error) {
	sortStr := "ORDER BY "

	switch sortBy {
	case "sum":
		sortStr = sortStr + "t.sum"
	case "date":
		sortStr = sortStr + "t.upd_time"
	default:
		return nil, errors.New("invalid sortBy")
	}

	if sortType == "desc" {
		sortStr = sortStr + " DESC"
	}

	sqlRow := fmt.Sprintf(`
SELECT t.order_id, t.service_id, t.sum, tt.type as "transaction_type", t.comment, t.upd_time
FROM public.transaction t
    LEFT JOIN public.transaction_type tt ON t.transaction_type_id = tt.id
WHERE t.user_id = $1
%s
OFFSET %d
LIMIT %d`, sortStr, offset, limit)

	rows, err := t.dbClient.Query(context.TODO(), sqlRow, userId)

	if err != nil {
		return nil, err
	}

	transactions := make([]model.Transaction, 0)

	for rows.Next() {
		var tr model.Transaction

		var orderId sql.NullString
		var serviceId sql.NullString
		var comment sql.NullString

		err = rows.Scan(&orderId, &serviceId, &tr.Sum, &tr.TransactionType, &comment, &tr.UpdTime)

		if err != nil {
			return nil, err
		}

		if orderId.Valid {
			tr.OrderId = &orderId.String
		} else {
			tr.OrderId = nil
		}

		if serviceId.Valid {
			tr.ServiceId = &serviceId.String
		} else {
			tr.ServiceId = nil
		}

		if comment.Valid {
			tr.Comment = &comment.String
		} else {
			tr.Comment = nil
		}

		transactions = append(transactions, tr)
	}

	return transactions, nil
}

package service

import (
	"context"
	"fmt"

	"github.com/avito-test/internal/config/logger"
	"github.com/avito-test/internal/model"
	"github.com/avito-test/internal/storage/repo"
	"github.com/sirupsen/logrus"
)

type TransactionService struct {
	repo repo.TransactionRepo
	log  *logrus.Logger
}

func NewTransactionService(repo repo.TransactionRepo) *TransactionService {
	return &TransactionService{
		repo: repo,
		log:  logger.GetLogger(),
	}
}

func (t *TransactionService) SaveTransaction(ctx context.Context, transaction model.Transaction) (int, error) {
	status, err := t.repo.SaveTransaction(transaction.OrderId, transaction.UserId, transaction.ServiceId, transaction.Sum, transaction.TransactionTypeId, transaction.Comment)

	if err != nil {
		t.log.WithFields(logrus.Fields{
			"error_message": err.Error(),
		}).Error(fmt.Sprintf("ERROR_%s", ctx.Value("requestId")))

		return 0, err
	}

	return status, nil
}

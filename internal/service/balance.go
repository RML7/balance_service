package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/avito-test/internal/config/logger"
	"github.com/avito-test/internal/model"
	"github.com/avito-test/internal/storage/repo"
	"github.com/sirupsen/logrus"
)

type BalanceService struct {
	BalanceNotFoundErr error

	repo repo.BalanceRepo
	log  *logrus.Logger
}

func NewBalanceService(repo repo.BalanceRepo) *BalanceService {
	return &BalanceService{
		BalanceNotFoundErr: errors.New("balance not found"),

		repo: repo,
		log:  logger.GetLogger(),
	}
}

func (b *BalanceService) AddBalance(ctx context.Context, transaction model.IncreaseBalanceTransaction) error {
	if err := b.repo.AddBalance(transaction.UserId, transaction.Sum, transaction.Comment); err != nil {
		b.log.WithFields(logrus.Fields{
			"error_message": err.Error(),
		}).Error(fmt.Sprintf("ERROR_%s", ctx.Value("requestId")))

		return err
	}

	return nil
}

func (b *BalanceService) GetBalanceByUserID(ctx context.Context, userId string) (float64, error) {
	balance, err := b.repo.GetBalanceByUserId(userId)

	if err != nil {
		b.log.WithFields(logrus.Fields{
			"error_message": err.Error(),
		}).Error(fmt.Sprintf("ERROR_%s", ctx.Value("requestId")))

		return 0, err
	}

	if balance == nil {
		return 0, b.BalanceNotFoundErr
	}

	return *balance, nil
}

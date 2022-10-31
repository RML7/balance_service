package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/avito-test/internal/config/logger"
	"github.com/avito-test/internal/storage/repo"
	"github.com/sirupsen/logrus"
)

type ReportService struct {
	repo repo.ReportRepo
	log  *logrus.Logger
}

func NewReportService(repo repo.ReportRepo) *ReportService {
	return &ReportService{
		repo: repo,
		log:  logger.GetLogger(),
	}
}

func (r *ReportService) CreateReport(ctx context.Context, dateFrom time.Time) error {
	rows, err := r.repo.GetReportRows(dateFrom)

	if err != nil {
		r.log.WithFields(logrus.Fields{
			"error_message": err.Error(),
		}).Error(fmt.Sprintf("ERROR_%s", ctx.Value("requestId")))

		return err
	}

	file, err := os.Create(fmt.Sprintf("static%sfile%s%s.csv", string(os.PathSeparator), string(os.PathSeparator), ctx.Value("requestId")))

	if err != nil {
		r.log.WithFields(logrus.Fields{
			"error_message": err.Error(),
		}).Error(fmt.Sprintf("ERROR_%s", ctx.Value("requestId")))

		return err
	}

	for _, row := range rows {
		_, err := file.WriteString(fmt.Sprintf("%s;%.6f\n", row.ServiceName, row.TotalSum))

		if err != nil {
			return err
		}
	}

	return nil
}

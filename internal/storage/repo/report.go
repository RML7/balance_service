package repo

import (
	"context"
	"time"

	"github.com/avito-test/internal/model"
	"github.com/avito-test/internal/storage/db"
)

type ReportRepo struct {
	dbClient db.Client
}

func NewReportRepo(dbClient db.Client) ReportRepo {
	return ReportRepo{dbClient: dbClient}
}

func (r *ReportRepo) GetReportRows(dateFrom time.Time) ([]model.ReportRow, error) {

	sqlRow := `
SELECT t.service_id, SUM(t.sum) as "total_sum"
FROM public.transaction t
WHERE t.transaction_type_id = 2 AND t.upd_time >= $1 AND t.upd_time < $2
GROUP BY t.service_id`

	rows, err := r.dbClient.Query(context.TODO(), sqlRow, dateFrom, dateFrom.AddDate(0, 1, 0))

	if err != nil {
		return nil, err
	}

	reportRows := make([]model.ReportRow, 0)

	for rows.Next() {
		var row model.ReportRow

		err = rows.Scan(&row.ServiceName, &row.TotalSum)

		if err != nil {
			return nil, err
		}

		reportRows = append(reportRows, row)
	}

	return reportRows, nil
}

package model

import "time"

type IncreaseBalanceTransaction struct {
	UserId  string
	Sum     float64
	Comment *string
}

type Transaction struct {
	UserId            string
	OrderId           *string
	ServiceId         *string
	Sum               float64
	TransactionTypeId int
	TransactionType   string
	Comment           *string
	UpdTime           time.Time
}

type GetTransactionsRequest struct {
	UserId   string
	Offset   int
	Limit    int
	SortBy   string
	SortType string
}

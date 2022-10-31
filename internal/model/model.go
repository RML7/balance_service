package model

type IncreaseBalanceTransaction struct {
	UserId  string
	Sum     float64
	Comment *string
}

type Transaction struct {
	UserId            string
	OrderId           string
	ServiceId         string
	Sum               float64
	TransactionTypeId int
	Comment           *string
}

package dto

import (
	"time"
)

type ApiError struct {
	Message string `json:"message"`
} //@name ApiError

type IncreaseBalanceRequest struct {
	UserId  *string  `json:"userId" validate:"required,uuid_rfc4122" example:"c806ce22-7ea3-4402-b979-9959746bb956" format:"uuid" binding:"required"`
	Sum     *float64 `json:"sum" validate:"required,numeric,gt=0" example:"53.68" format:"numeric" binding:"required"`
	Comment *string  `jsom:"comment" example:"Зачисление денежных средств на баланс" format:"string"`
} //@name IncreaseBalanceRequest

type GetBalanceResponse struct {
	Balance float64 `json:"balance" example:"53.68" format:"numeric"`
} //@name GetBalanceResponse

type SaveTransactionRequest struct {
	UserId            *string  `json:"userId" validate:"required,uuid_rfc4122" example:"e8c49cf0-d984-4ed8-a37c-2d60f74c7fe5" format:"uuid"`
	OrderId           *string  `json:"orderId" validate:"required,uuid_rfc4122" example:"6c87959d-aa88-4f51-932b-ff70563ad87a" format:"uuid"`
	ServiceId         *string  `json:"serviceId" validate:"required,uuid_rfc4122" example:"15aa9f91-c8f7-40e4-9108-d45891c10444" format:"uuid"`
	Sum               *float64 `json:"sum" validate:"required,numeric,gt=0" example:"345" format:"numeric"`
	TransactionTypeId *int     `json:"transactionType" validate:"required,numeric,oneof=1 2 3" example:"1" format:"integer" enums:"1,2,3"`
	Comment           *string  `jsom:"comment" example:"Резервация денежных средств" format:"string"`
} //@name SaveTransactionRequest

type SaveTransactionResponse struct {
	Status int `json:"status" example:"1"`
} //@name SaveTransactionResponse

type GetTransactionsRequest struct {
	UserId       *string `json:"userId" validate:"required,uuid_rfc4122"`
	Page         *int    `json:"page" validate:"required,min=1"`
	ItemsPerPage *int    `json:"itemsPerPage" validate:"omitempty,min=1"`
	SortBy       *string `json:"sortBy" validate:"omitempty,oneof=sum date"`
	SortType     *string `json:"sortType" validate:"omitempty,oneof=asc desc"`
}

type GetTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	UserId            *string   `json:"user_id,omitempty" swaggerignore:"true"`
	OrderId           *string   `json:"order_id,omitempty" example:"6c87959d-aa88-4f51-932b-ff70563ad87b"`
	ServiceId         *string   `json:"service_id,omitempty" example:"15aa9f91-c8f7-40e4-9108-d45891c10444"`
	Sum               float64   `json:"sum" example:"1000"`
	TransactionTypeId *int      `json:"transaction_type_id,omitempty" swaggerignore:"true"`
	TransactionType   string    `json:"transaction_type" example:"Резервация подтверждена, средства списаны, оплата прошла"`
	Comment           *string   `json:"comment,omitempty" example:"оплата подтверждена"`
	UpdTime           time.Time `json:"date" example:"2022-11-01T16:37:52.717392Z"`
}

type CreateReportRequest struct {
	Year  *int `validate:"required,min=2022,max=2100"`
	Month *int `validate:"required,oneof=1 2 3 4 5 6 7 8 9 10 11 12"`
}

type CreateReportResponse struct {
	URL string `json:"url"`
}

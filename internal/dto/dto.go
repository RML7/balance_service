package dto

type ApiError struct {
	Message string `json:"message"`
} //@name ApiError

type IncreaseBalanceRequest struct {
	UserId  *string  `json:"userId" validate:"required,uuid" example:"c806ce22-7ea3-4402-b979-9959746bb956" format:"uuid" binding:"required"`
	Sum     *float64 `json:"sum" validate:"required,numeric,gt=0" example:"53.68" format:"numeric" binding:"required"`
	Comment *string  `jsom:"comment" validate:"min=1" example:"Зачисление денежных средств на баланс" format:"string"`
} //@name IncreaseBalanceRequest

type GetBalanceResponse struct {
	Balance float64 `json:"balance" example:"53.68" format:"numeric"`
} //@name GetBalanceResponse

type SaveTransactionRequest struct {
	UserId            *string  `json:"userId" validate:"required,uuid" example:"e8c49cf0-d984-4ed8-a37c-2d60f74c7fe5" format:"uuid"`
	OrderId           *string  `json:"orderId" validate:"required,uuid" example:"6c87959d-aa88-4f51-932b-ff70563ad87a" format:"uuid"`
	ServiceId         *string  `json:"serviceId" validate:"required,uuid" example:"15aa9f91-c8f7-40e4-9108-d45891c10444" format:"uuid"`
	Sum               *float64 `json:"sum" validate:"required,numeric,gt=0" example:"345" format:"numeric"`
	TransactionTypeId *int     `json:"transactionType" validate:"required,numeric,oneof=1 2 3" example:"1" format:"integer" enums:"1,2,3"`
} //@name SaveTransactionRequest

type SaveTransactionResponse struct {
	Status int `json:"status" example:"1"`
} //@name SaveTransactionResponse

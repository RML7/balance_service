package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/avito-test/internal/config/logger"
	valid "github.com/avito-test/internal/config/validator"
	"github.com/avito-test/internal/dto"
	"github.com/avito-test/internal/model"
	"github.com/avito-test/internal/service"
	"github.com/avito-test/internal/storage/db"
	"github.com/avito-test/internal/storage/repo"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type httpServer struct {
	InternalServerError error

	validator          *validator.Validate
	log                *logrus.Logger
	balanceService     *service.BalanceService
	transactionService *service.TransactionService
	reportService      *service.ReportService
}

func NewHttpServer() *httpServer {
	log := logger.GetLogger()

	dbClient, err := db.NewDbPool(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}

	balanceRepo := repo.NewBalanceRepo(dbClient)
	transactionRepo := repo.NewTransactionRepo(dbClient)
	reportRepo := repo.NewReportRepo(dbClient)

	return &httpServer{
		InternalServerError: errors.New("internal server error"),

		validator:          valid.GetValidator(),
		log:                log,
		balanceService:     service.NewBalanceService(balanceRepo),
		transactionService: service.NewTransactionService(transactionRepo),
		reportService:      service.NewReportService(reportRepo),
	}
}

// HandleIncreaseBalance
// @summary увеличение баланса
// @tags balance
// @description Метод для увеличения баланса
// @accept json
// @produce json
// @param IncreaseBalanceRequest body dto.IncreaseBalanceRequest true "userId - id пользователя (UUID)<br>sum - сумма пополнения (больше 0)<br> comment - комментарий (опционально)"
// @success 200 "В случае успешного добавления денег к балансу возвращается статус 200"
// @failure 400 {object} dto.ApiError "В случае если запрос не валидный возвращается статус 400 и тело ответа"
// @router /balance [post]
func (s *httpServer) HandleIncreaseBalance(w http.ResponseWriter, r *http.Request) {
	var request dto.IncreaseBalanceRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: "invalid request body"})
		return
	}

	if ok, validationMessage, err := s.isValidRequest(r.Context(), s.validator, request); err != nil {
		if errors.Is(err, s.InternalServerError) {
			s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: err.Error()})
		} else {
			s.log.WithFields(logrus.Fields{
				"error_message": err.Error(),
			}).Error(fmt.Sprintf("%s_ERROR", r.Context().Value("requestId")))
		}

		return
	} else {
		if !ok {
			s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: validationMessage})
			return
		}
	}

	transaction := model.IncreaseBalanceTransaction{
		UserId:  *request.UserId,
		Sum:     *request.Sum,
		Comment: request.Comment,
	}

	if err := s.balanceService.AddBalance(r.Context(), transaction); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: s.InternalServerError.Error()})
	}
}

// HandleGetBalance
// @summary получение баланса по userId
// @tags balance
// @description Метод для получения баланса по userId
// @accept json
// @produce json
// @param userId path string true "id пользователя" Format(uuid) example(b2b9a788-55fb-11ed-bdc3-0242ac120002)
// @success 200 {object} dto.GetBalanceResponse "В случае если баланс найден"
// @failure 400 {object} dto.ApiError "В случае если невалидный userId"
// @failure 404 {object} dto.ApiError "В случае если баланс не найден по userId"
// @router /balance/{userId} [get]
func (s *httpServer) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if err := s.validator.Var(params["userId"], "uuid"); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: "parameter userId should be uuid"})
		return
	}

	balance, err := s.balanceService.GetBalanceByUserID(r.Context(), params["userId"])

	if err != nil {
		if errors.Is(err, s.balanceService.BalanceNotFoundErr) {
			s.sendJsonResponse(r.Context(), w, http.StatusNotFound, dto.ApiError{Message: err.Error()})
		} else {
			s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: s.InternalServerError.Error()})
		}

		return
	}

	s.sendJsonResponse(r.Context(), w, http.StatusOK, dto.GetBalanceResponse{Balance: balance})
}

// HandleTransaction
// @summary Метод для обработки транзакции
// @tags transaction
// @description Метод для обработки транзакции. Для резервации денег со счета в теле запроса поле "transactionType" = 1. Для признания выручки и подтверждения списания средств с баланса "transactionType" = 2. В случае отмены резервации и возврата средств на основной баланс "transactionType" = 3.
// @accept json
// @produce json
// @param SaveTransactionRequest body dto.SaveTransactionRequest true "orderId - id заказа (UUID).<br> serviceId - id услуги (UUID).<br> userId - id пользователя (UUID).<br> sum - сумма транзакции (больше 0).<br> transactionType - тип транзакции (enum(1, 2, 3)).<br> comment - комментарий (опционально)"
// @success 200 {object} dto.SaveTransactionResponse "Воможные статусы:<br> 1 - добавление/обновление произошло успешно.<br> 2 - попытка резервации ("transactionType" = 1), резервация с соответсвующими orderId, userId, serviceId уже существует, транзакция резервации не добавлена.<br> 3 - попытка резервации ("transactionType" = 1), на балансе недостаточно средств, транзакция резервации не добавлена.<br> 4 - попытка признания выручки ("transactionType" = 2), транзакция с соответсвующими orderId, userId, serviceId не найдена, ошибка.<br> 5 - попытка признания выручки ("transactionType" = 2), транзакция с соответсвующими orderId, userId, serviceId найдена и выручка уже списана ранее, ошибка.<br> 6 - попытка признания выручки ("transactionType" = 2), транзакция с соответсвующими orderId, userId, serviceId найдена и уже была отклонена ранее, ошибка.<br> 7 - попытка отмены резервации ("transactionType" = 3), транзакция с соответсвующими orderId, userId, serviceId не найдена, ошибка.<br> 8 - попытка отмены резервации ("transactionType" = 3), транзакция с соответсвующими orderId, userId, serviceId найдена и уже была отменена ранее, деньги были возвращены, ошибка.<br> 9 - попытка отмены резервации ("transactionType" = 3), транзакция с соответсвующими orderId, userId, serviceId найдена и уже была подтверждена ранее, деньги были списаны, ошибка.<br> 10 - баланс пользователя не найден, ошибка"
// @router /transaction [post]
func (s *httpServer) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	var request dto.SaveTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: "invalid request body"})
		return
	}

	if ok, validationMessage, err := s.isValidRequest(r.Context(), s.validator, request); err != nil {
		if errors.Is(err, s.InternalServerError) {
			s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: err.Error()})
		} else {
			s.log.WithFields(logrus.Fields{
				"error_message": err.Error(),
			}).Error(fmt.Sprintf("%s_ERROR", r.Context().Value("requestId")))
		}

		return
	} else {
		if !ok {
			s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: validationMessage})
			return
		}
	}

	transaction := model.Transaction{
		UserId:            *request.UserId,
		OrderId:           request.OrderId,
		ServiceId:         request.ServiceId,
		Sum:               *request.Sum,
		TransactionTypeId: *request.TransactionTypeId,
		Comment:           request.Comment,
	}

	if status, err := s.transactionService.SaveTransaction(r.Context(), transaction); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: s.InternalServerError.Error()})
	} else {
		s.sendJsonResponse(r.Context(), w, http.StatusOK, dto.SaveTransactionResponse{Status: status})
	}
}

// HandleGetTransactions
// @summary Получение списка транзакций пользователя
// @tags transaction
// @description Метод получение списка транзакций пользователя
// @accept json
// @produce json
// @param userId query string true "id пользователя" Format(uuid) example(b2b9a788-55fb-11ed-bdc3-0242ac120002)
// @param page query integer true "номер страницы" example(1) minimum(1)
// @param itemsPerPage query integer false "количество записей на странице" example(1) minimum(1) default(10)
// @param sortBy query string false "поле по которому надо сортировать" example(sum) default(date) enums(date, sum)
// @param sortType query string false "тип сортировки" example(asc) default(desc) enums(asc, desc)
// @success 200 {object} dto.GetTransactionsResponse
// @router /transaction [get]
func (s *httpServer) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.GetTransactionsRequest

	queryParams := r.URL.Query()

	if userIds, ok := queryParams["userId"]; ok {
		requestDto.UserId = &userIds[0]
	}

	if pages, ok := queryParams["page"]; ok {
		if i, err := strconv.Atoi(pages[0]); err != nil {
			s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: "parameter page should be integer"})
			return
		} else {
			requestDto.Page = &i
		}
	}

	if itemsPerPageArr, ok := queryParams["itemsPerPage"]; ok {
		if i, err := strconv.Atoi(itemsPerPageArr[0]); err != nil {
			s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: "parameter itemsPerPage should be integer"})
			return
		} else {
			requestDto.ItemsPerPage = &i
		}
	}

	if sortByArr, ok := queryParams["sortBy"]; ok {
		requestDto.SortBy = &sortByArr[0]
	}

	if sortTypeArr, ok := queryParams["sortType"]; ok {
		requestDto.SortType = &sortTypeArr[0]
	}

	if ok, validationMessage, err := s.isValidRequest(r.Context(), s.validator, requestDto); err != nil {
		if errors.Is(err, s.InternalServerError) {
			s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: err.Error()})
		} else {
			s.log.WithFields(logrus.Fields{
				"error_message": err.Error(),
			}).Error(fmt.Sprintf("%s_ERROR", r.Context().Value("requestId")))
		}

		return
	} else {
		if !ok {
			s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: validationMessage})
			return
		}
	}

	var request model.GetTransactionsRequest

	if requestDto.ItemsPerPage == nil {
		request.Limit = 10
	} else {
		request.Limit = *requestDto.ItemsPerPage
	}

	request.UserId = *requestDto.UserId
	request.Offset = (*requestDto.Page - 1) * request.Limit

	if requestDto.SortBy == nil {
		request.SortBy = "date"
	} else {
		request.SortBy = *requestDto.SortBy
	}

	if requestDto.SortType == nil {
		request.SortType = "desc"
	} else {
		request.SortType = *requestDto.SortType
	}

	if transactions, err := s.transactionService.GetTransactions(r.Context(), request); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: "internal server error"})
	} else {
		var response dto.GetTransactionsResponse

		response.Transactions = make([]dto.Transaction, 0)

		for _, tr := range transactions {
			response.Transactions = append(response.Transactions, dto.Transaction{
				OrderId:         tr.OrderId,
				ServiceId:       tr.ServiceId,
				TransactionType: tr.TransactionType,
				Sum:             tr.Sum,
				Comment:         tr.Comment,
				UpdTime:         tr.UpdTime,
			})
		}

		s.sendJsonResponse(r.Context(), w, http.StatusOK, response)
	}
}

// HandleCreateReport
// @summary Создание отчета для бухгалтерии
// @tags report
// @description Метод создание отчета для бухгалтерии. Возвращает ссылку на файл
// @accept json
// @produce json
// @param CreateReportRequest body dto.CreateReportRequest true "year - год отчета (2022 <=year <= 2100)<br>month - месяц отчета"
// @success 200 {object} dto.CreateReportResponse
// @router /report [post]
func (s *httpServer) HandleCreateReport(w http.ResponseWriter, r *http.Request) {
	var request dto.CreateReportRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: "invalid request body"})
		return
	}

	if ok, validationMessage, err := s.isValidRequest(r.Context(), s.validator, request); err != nil {
		if errors.Is(err, s.InternalServerError) {
			s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: err.Error()})
		} else {
			s.log.WithFields(logrus.Fields{
				"error_message": err.Error(),
			}).Error(fmt.Sprintf("%s_ERROR", r.Context().Value("requestId")))
		}

		return
	} else {
		if !ok {
			s.sendJsonResponse(r.Context(), w, http.StatusBadRequest, dto.ApiError{Message: validationMessage})
			return
		}
	}

	t := time.Date(*request.Year, time.Month(*request.Month), 0, 0, 0, 0, 0, time.UTC)

	if err := s.reportService.CreateReport(r.Context(), t); err != nil {
		s.sendJsonResponse(r.Context(), w, http.StatusInternalServerError, dto.ApiError{Message: "internal server error"})
	} else {
		s.sendJsonResponse(r.Context(), w, http.StatusOK, dto.CreateReportResponse{
			URL: fmt.Sprintf("%s/%s.csv", "http://localhost:8000/report", r.Context().Value("requestId")),
		})
	}
}

func (s *httpServer) isValidRequest(ctx context.Context, v *validator.Validate, requestBody any) (bool, string, error) {
	ok := true
	var validationMessage string

	if err := v.Struct(requestBody); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "required":
				switch err.Type().Kind() {
				case reflect.Pointer:
					validationMessage = fmt.Sprintf("field %s missing", err.Field())
				default:
					s.log.WithFields(logrus.Fields{
						"error_message": fmt.Sprintf("unexpected field type when validating required tag: %s", err.Type()),
					}).Error(fmt.Sprintf("%s_ERROR", ctx.Value("requestId")))

					return false, "", s.InternalServerError
				}
			case "gt":
				switch err.Type().Kind() {
				case reflect.Float64:
					validationMessage = fmt.Sprintf("field %s should be > %s", err.Field(), err.Param())
				default:
					s.log.WithFields(logrus.Fields{
						"error_message": fmt.Sprintf("unexpected field type when validating gt tag: %s", err.Type()),
					}).Error(fmt.Sprintf("%s_ERROR", ctx.Value("requestId")))

					return false, "", s.InternalServerError
				}
			case "uuid_rfc4122":
				validationMessage = fmt.Sprintf("field %s should be uuid", err.Field())
			case "oneof":
				validationMessage = fmt.Sprintf("field %s should be in [%s]", err.Field(), err.Param())
			case "min":
				switch err.Type().Kind() {
				case reflect.Int:
					validationMessage = fmt.Sprintf("field %s should be >= %s", err.Field(), err.Param())
				default:
					s.log.WithFields(logrus.Fields{
						"error_message": fmt.Sprintf("unexpected field type when validating min tag: %s", err.Type()),
					}).Error(fmt.Sprintf("%s_ERROR", ctx.Value("requestId")))

					return false, "", s.InternalServerError
				}
			case "max":
				switch err.Type().Kind() {
				case reflect.Int:
					validationMessage = fmt.Sprintf("field %s should be <= %s", err.Field(), err.Param())
				default:
					s.log.WithFields(logrus.Fields{
						"error_message": fmt.Sprintf("unexpected field type when validating max tag: %s", err.Type()),
					}).Error(fmt.Sprintf("%s_ERROR", ctx.Value("requestId")))

					return false, "", s.InternalServerError
				}
			default:
				s.log.WithFields(logrus.Fields{
					"error_message": fmt.Sprintf("unexpected validation tag: %s", err.Tag()),
				}).Error(fmt.Sprintf("%s_ERROR", ctx.Value("requestId")))

				return false, "", s.InternalServerError
			}

			ok = false
			break
		}
	}

	return ok, validationMessage, nil
}

func (s *httpServer) sendJsonResponse(ctx context.Context, w http.ResponseWriter, status int, v interface{}) {
	w.WriteHeader(status)

	response, err := json.Marshal(v)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"error_message": err.Error(),
		}).Error(fmt.Sprintf("%s_ERROR", ctx.Value("requestId")))

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(response); err != nil {
		s.log.Fatal(err.Error())
	}
}

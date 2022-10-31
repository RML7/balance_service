package main

import (
	"log"
	"net/http"

	_ "github.com/avito-test/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/avito-test/internal/config/middleware"
	"github.com/avito-test/internal/server"
	"github.com/gorilla/mux"

	swagger "github.com/swaggo/http-swagger"
)

// @title Balance Service
// @version 2.0

// @host localhost:8000
func main() {
	router := mux.NewRouter()
	router.Use(middleware.ResponseHeaders, middleware.RequestID, middleware.Logging)

	httpServer := server.NewHttpServer()

	router.HandleFunc("/balance/{userId}", httpServer.HandleGetBalance).Methods(http.MethodGet)
	router.HandleFunc("/balance", httpServer.HandleIncreaseBalance).Methods(http.MethodPost)

	router.HandleFunc("/transaction", httpServer.HandleGetTransactions).Methods(http.MethodGet)
	router.HandleFunc("/transaction", httpServer.HandleTransaction).Methods(http.MethodPost)

	router.HandleFunc("/report", httpServer.HandleCreateReport).Methods(http.MethodPost)

	router.PathPrefix("/swagger").Handler(swagger.WrapHandler).Methods(http.MethodGet)

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal(err.Error())
	}
}

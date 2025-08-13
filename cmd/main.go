package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"public_library/internal/book"
	"public_library/internal/db"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	_ "public_library/docs"
)

// @title Public Library API
// @version 1.0
// @description A minimal REST API for managing books in a fictional public library
// @host localhost:8080
// @BasePath /api/v1/
// @schemes http
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load config from YAML file
	cfg, err := db.LoadConfigFromYAML("config/config.yaml")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	dbConn := db.InitConnection(cfg, logger)
	repo := book.NewRepository(dbConn)
	handler := book.NewHandler(repo, logger)

	// RESTful routes
	router := mux.NewRouter()
	v1 := router.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	v1.HandleFunc("/books/list", handler.GetBooks).Methods("POST")
	v1.HandleFunc("/books/create", handler.CreateBook).Methods("POST")
	v1.HandleFunc("/books/{id}", handler.GetBookByID).Methods("GET")
	v1.HandleFunc("/books/{id}", handler.UpdateBook).Methods("PUT")
	v1.HandleFunc("/books/{id}", handler.DeleteBook).Methods("DELETE")

	v1.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	logger.Info("Starting server", zap.String("addr", ":8080"))
	log.Fatal(http.ListenAndServe(":8080", router))
}

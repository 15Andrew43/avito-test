package main

import (
	"fmt"
	"log"
	"net/http"
	"tender-service/api/handlers"
	"tender-service/config"
	"tender-service/internal/repository"
	"tender-service/internal/service"

	"database/sql"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.LoadConfig()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tenderRepo := repository.NewTenderRepository(db)
	userRepo := repository.NewUserRepository(db)
	bidRepo := repository.NewBidRepository(db)

	userService := service.NewUserService(userRepo)
	tenderService := service.NewTenderService(tenderRepo, userService)
	bidService := service.NewBidService(bidRepo, tenderRepo, userRepo)

	tenderHandler := handlers.NewTenderHandler(tenderService, userService)
	bidHandler := handlers.NewBidHandler(bidService)

	router := mux.NewRouter()
	router.HandleFunc("/api/ping", handlers.PingHandler).Methods("GET")
	router.HandleFunc("/api/tenders", tenderHandler.GetTenders).Methods("GET")
	router.HandleFunc("/api/tenders/new", tenderHandler.CreateTender).Methods("POST")
	router.HandleFunc("/api/tenders/my", tenderHandler.GetUserTenders).Methods("GET")

	router.HandleFunc("/api/tenders/{tenderId}/status", tenderHandler.GetTenderStatus).Methods("GET")
	router.HandleFunc("/api/tenders/{tenderId}/status", tenderHandler.UpdateTenderStatus).Methods("PUT")
	router.HandleFunc("/api/tenders/{tenderId}/edit", tenderHandler.EditTender).Methods("PATCH")
	router.HandleFunc("/api/tenders/{tenderId}/rollback/{version}", tenderHandler.RollbackTenderVersion).Methods("POST")

	router.HandleFunc("/api/bids/new", bidHandler.CreateBid).Methods("POST")
	router.HandleFunc("/api/bids/my", bidHandler.GetUserBids).Methods("GET")
	router.HandleFunc("/api/bids/{tenderId}/list", bidHandler.GetBidsByTenderID).Methods("GET")
	router.HandleFunc("/api/bids/{bidId}/status", bidHandler.GetBidStatus).Methods("GET")
	router.HandleFunc("/api/bids/{bidId}/status", bidHandler.UpdateBidStatus).Methods("PUT")
	router.HandleFunc("/api/bids/{bidId}/edit", bidHandler.EditBid).Methods("PATCH")
	router.HandleFunc("/api/bids/{bidId}/feedback", bidHandler.SubmitBidFeedback).Methods("PUT")

	log.Printf("Server running at %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}

package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"wallet/internal/config"
	"wallet/internal/delivery/http/handler"
	"wallet/internal/delivery/http/router"
	"wallet/internal/models"
	"wallet/internal/repository"
	"wallet/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	server *http.Server
	db     *gorm.DB
}

func New(addr string) *Server {
	config.InitFromFile()
	db, err := initDatabase()
	if err != nil {
		panic(err)
	}

	playerRepo := repository.NewPlayerRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	walletService := service.NewWalletService(db, playerRepo, transactionRepo)
	trxHandler := handler.NewTransactionHandler(walletService)
	playerHandler := handler.NewPlayerHandler(walletService)

	r := router.New(trxHandler, playerHandler)
	return &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           r,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
		db: db,
	}

}
func (s *Server) Run() error {
	log.Printf("starting server")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		log.Fatalln(err)
	}

	return sqlDB.Close()
}

func initDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		config.Config.Postgres.Host,
		config.Config.Postgres.Port,
		config.Config.Postgres.User,
		config.Config.Postgres.Password,
		config.Config.Postgres.Database,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	err = db.AutoMigrate(&models.Player{}, &models.Transaction{})
	if err != nil {
		return nil, err
	}

	return db, nil

}

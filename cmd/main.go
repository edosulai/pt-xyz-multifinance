package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edosulai/pt-xyz-multifinance/internal/handler"
	"github.com/edosulai/pt-xyz-multifinance/internal/repo"
	"github.com/edosulai/pt-xyz-multifinance/internal/usecase"
	"github.com/edosulai/pt-xyz-multifinance/pkg/config"
	"github.com/edosulai/pt-xyz-multifinance/pkg/database"
	"github.com/edosulai/pt-xyz-multifinance/pkg/logger"
	"github.com/edosulai/pt-xyz-multifinance/pkg/middleware"
	pb "github.com/edosulai/pt-xyz-multifinance/proto/gen/go/xyz/multifinance/v1"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Initialize logger
	if err := logger.InitLogger(cfg.Logging.Level, cfg.Logging.Encoding, cfg.Logging.OutputPaths); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	log := logger.GetLogger()
	defer log.Sync()

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize dependencies with wrapped database
	wrappedDB := &database.DB{DB: db}
	userRepo := repo.NewUserRepository(wrappedDB)
	loanRepo := repo.NewLoanRepository(wrappedDB)

	userUseCase, err := usecase.NewUserUseCase(userRepo, cfg.JWT.SecretKey, cfg.JWT.Expiration)
	if err != nil {
		log.Fatal("Failed to initialize user use case", zap.Error(err))
	}
	loanUseCase := usecase.NewLoanUseCase(loanRepo, userRepo)

	authInterceptor := middleware.NewAuthInterceptor(cfg.JWT.SecretKey)

	// Create channels for graceful shutdown
	grpcShutdown := make(chan struct{})
	httpShutdown := make(chan struct{})
	// Start gRPC server
	grpcServer := initGRPCServer(cfg, log, userUseCase, loanUseCase, authInterceptor, grpcShutdown)

	// Start HTTP server with gRPC-Gateway
	httpServer := initHTTPServer(cfg, log, userUseCase, httpShutdown)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	// Shutdown both servers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP server shutdown error", zap.Error(err))
	}
	<-httpShutdown

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	<-grpcShutdown

	log.Info("Servers exited properly")
}

func initGRPCServer(cfg *config.Config, log *zap.Logger, userUseCase usecase.UserUseCase, loanUseCase usecase.LoanUseCase, authInterceptor *middleware.AuthInterceptor, shutdown chan struct{}) *grpc.Server {
	// Initialize gRPC server with middleware
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryServerInterceptor()),
	)
	// Register services
	userHandler := handler.NewUserHandler(userUseCase, log)
	loanHandler := handler.NewLoanHandler(loanUseCase, log)

	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterLoanServiceServer(grpcServer, loanHandler)
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		defer close(shutdown)

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
		if err != nil {
			log.Fatal("Failed to listen for gRPC", zap.Error(err))
		}

		log.Info("gRPC server started", zap.Int("port", cfg.Server.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server error", zap.Error(err))
		}
	}()

	return grpcServer
}

func initHTTPServer(cfg *config.Config, log *zap.Logger, userUseCase usecase.UserUseCase, shutdown chan struct{}) *http.Server {
	// Initialize gRPC-Gateway
	ctx := context.Background()
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := pb.RegisterUserServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort),
		opts,
	); err != nil {
		log.Fatal("Failed to register user service handler", zap.Error(err))
	}

	// Register loan service handler
	if err := pb.RegisterLoanServiceHandlerFromEndpoint(
		ctx,
		gwmux,
		fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort),
		opts,
	); err != nil {
		log.Fatal("Failed to register loan service handler", zap.Error(err))
	} // Initialize router with both gRPC-Gateway and HTTP handlers
	router := mux.NewRouter()

	// Serve Swagger UI with combined swagger files
	swaggerFiles := []string{
		"proto/gen/openapiv2/proto/user.swagger.json",
		"proto/gen/openapiv2/proto/loan.swagger.json",
	}
	swaggerHandler := handler.SwaggerHandler(swaggerFiles)
	router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", swaggerHandler))

	// Add CORS middleware for Swagger UI
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
			if r.Method == "OPTIONS" {
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	// Add additional HTTP routes first
	httpHandler := handler.NewHTTPHandler(log)
	httpHandler.RegisterHTTPRoutes(router)

	// Then serve gRPC-Gateway API endpoints
	router.PathPrefix("/v1/").Handler(gwmux)

	// Configure HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
	go func() {
		defer close(shutdown)

		log.Info("HTTP server started", zap.Int("port", cfg.Server.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", zap.Error(err))
		}
	}()
	return srv
}

func initDatabase(cfg *config.Config) (*sql.DB, error) {
	// Construct DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	// Connect to database with retries
	var db *sql.DB
	var err error
	maxRetries := 5

	for retry := 0; retry < maxRetries; retry++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			// Test the connection
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = db.PingContext(ctx)
			cancel()

			if err == nil {
				break
			}
		}

		log := logger.GetLogger()
		log.Warn("Failed to connect to database, retrying...",
			zap.Error(err),
			zap.Int("attempt", retry+1),
			zap.Int("maxRetries", maxRetries))

		if db != nil {
			db.Close()
		}

		// Wait before retrying, with exponential backoff
		time.Sleep(time.Second * time.Duration(1<<uint(retry)))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	return db, nil
}

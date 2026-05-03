package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/IsFariza/ap2-Message-Queue/doctor-service/internal/event"
	"github.com/IsFariza/ap2-Message-Queue/doctor-service/internal/repository"
	grpcHandler "github.com/IsFariza/ap2-Message-Queue/doctor-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Message-Queue/doctor-service/internal/usecase"
	pb "github.com/IsFariza/ap2-Message-Queue/doctor-service/proto"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
)

func main() {
	// 1. Load Configurations
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	natsURL := os.Getenv("NATS_URL")
	grpcPort := os.Getenv("PORT")
	if grpcPort == "" {
		grpcPort = "50051" // Fallback default
	}

	// 2. Connect to PostgreSQL
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL")

	// 3. Run Migrations
	runMigrations(dbURL)

	// 4. Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Successfully connected to NATS")

	// 5. Wire up Layers (Dependency Injection)
	repo := repository.NewDoctorRepository(db)
	pub := event.NewDoctorPublisher(nc)
	uc := usecase.NewDoctorUseCase(repo, pub)
	handler := grpcHandler.NewDoctorHandler(uc)

	// 6. Setup gRPC Server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	server := grpc.NewServer()
	pb.RegisterDoctorServiceServer(server, handler)

	// 7. Graceful Shutdown Setup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Doctor Service listening on %s", grpcPort)
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("Failed to serve grpc: %v", err)
		}
	}()

	// Wait for interruption signal
	<-ctx.Done()
	log.Println("Shutting down Doctor Service gracefully...")

	server.GracefulStop()
	log.Println("Doctor Service stopped safely.")
}

func runMigrations(dbURL string) {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("Migration init error: %v", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Database is up to date")
		} else {
			log.Fatalf("Migration failed: %v", err)
		}
	} else {
		log.Println("Migrations applied successfully")
	}
}

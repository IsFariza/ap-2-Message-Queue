package main

import (
	"database/sql"
	"log"
	"net"
	"os"

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
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println(".env not found")
	}

	dbURL := os.Getenv("DATABASE_URL")

	natsURL := os.Getenv("NATS_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to postgreSQL")

	runMigrations(dbURL)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Successfully connected to NATS")

	repo := repository.NewDoctorRepository(db)
	pub := event.NewDoctorPublisher(nc)
	uc := usecase.NewDoctorUseCase(repo, pub)
	handler := grpcHandler.NewDoctorHandler(uc)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen on port 50051: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterDoctorServiceServer(s, handler)

	log.Println("Doctor Service is listening on 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve grpc: %v", err)
	}
}

func runMigrations(dbURL string) {
	m, err := migrate.New("file://../../migrations", dbURL)
	if err != nil {
		log.Fatalf("Migration init error: %v", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No changes in database")
		} else {
			log.Fatalf("Migration failed: %v", err)
		}
	} else {
		log.Println("Migrations applied successfully")
	}
}

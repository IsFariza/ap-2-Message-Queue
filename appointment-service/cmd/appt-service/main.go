package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	// Updated paths to match ap2-Message-Queue
	pb "github.com/IsFariza/ap2-Message-Queue/appointment-service/appointment_proto"
	"github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/client"
	"github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/event"
	"github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/repository"
	transportgrpc "github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/transport/grpc"
	"github.com/IsFariza/ap2-Message-Queue/appointment-service/internal/usecase"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	// 1. Load Configurations
	dbURL := os.Getenv("DATABASE_URL") // e.g., postgres://user:pass@localhost:5432/db
	natsURL := os.Getenv("NATS_URL")
	grpcPort := os.Getenv("PORT")
	doctorAddr := os.Getenv("DOCTOR_ADDR")

	// 2. Connect to PostgreSQL
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Postgres Connection Error: %v", err)
	}
	defer db.Close()

	// 3. Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("NATS Connection Error: %v", err)
	}
	defer nc.Close()

	// 4. Initialize Doctor gRPC Client
	conn, err := grpc.NewClient(
		doctorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Doctor Service client: %v", err)
	}
	defer conn.Close()

	// 5. Wire up the Layers (Dependency Injection)
	docClient := client.NewDoctorClient(conn)
	repo := repository.NewAppointmentRepository(db)
	pub := event.NewAppointmentPublisher(nc) // Using your new lowercase struct via constructor

	// Pass repo, docClient, and publisher to the UseCase
	apptUsecase := usecase.NewAppointmentUsecase(repo, docClient, pub)
	handler := transportgrpc.NewAppointmentHandler(apptUsecase)

	// 6. Start gRPC Server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAppointmentServiceServer(server, handler)

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Appointment service starting on port %s", grpcPort)
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC Server Error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	server.GracefulStop()
	log.Println("Service stopped safely.")
}

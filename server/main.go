package main

import (
	"context"
	pb "converter/converter"
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


type Server struct {
	pb.ConverterServer
	db *sql.DB
}

func main() {
	db, err := sql.Open("postgres", "postgresql://username:password@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("Failed to listen on port 9000: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterConverterServer(grpcServer, &Server{db : db})
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
	}
	log.Println("Server started running on port 9000")
}
func (s *Server) ConvertCurrency(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	amount := req.Amount
	fromCurrency := req.FromCurrency
	toCurrency := req.ToCurrency
	var fromValue float64
	var toValue float64

	err := s.db.QueryRow("SELECT conversion_value FROM currency_table WHERE currency = $1", fromCurrency).Scan(&fromValue)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Conversion value for %s is not available", fromCurrency)
	}

	err = s.db.QueryRow("SELECT conversion_value FROM currency_table WHERE currency = $1", toCurrency).Scan(&toValue)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Conversion value for %s is not available", toCurrency)
	}

	convertedAmount := amount * fromValue / toValue
	return &pb.ConvertResponse{Amount: convertedAmount}, nil
}




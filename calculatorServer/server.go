package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"microservices/calculatorpb"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Multiplication(ctx context.Context, req *calculatorpb.MultiplicationRequest) (*calculatorpb.MultiplicationResponse, error) {
	fmt.Println("Start Multiplication")
	nb1 := req.GetNumbers().FirstNb
	nb2 := req.GetNumbers().SecondNb
	result := nb1 * nb2
	res := &calculatorpb.MultiplicationResponse{
		Result: result,
	}
	return res, nil
}

func (*server) Add(stream calculatorpb.CalculatorService_AddServer) error {
	fmt.Println("Start Add")
	var sum int32
	for {
		time.Sleep(500 * time.Millisecond)
		req, err := stream.Recv()
		sum += req.GetNumber()
		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.AddResponse{
				Result: sum,
			})
		}
		if err != nil {
			log.Fatalf("Error client stream: %v", err)
		}
	}
}

func (*server) Fibonacci(req *calculatorpb.FibonacciRequest, stream calculatorpb.CalculatorService_FibonacciServer) error {
	fmt.Println("Start Fibonacci")
	until := req.GetIterations()
	if until < 0 {
		return status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received negative iterations limit : %v", until),
		)
	}
	oldFibonacci := int32(0)
	fibonacci := int32(1)
	for i := int32(0); i < until; i++ {
		time.Sleep(500 * time.Millisecond)
		saveFibo := fibonacci
		fibonacci = fibonacci + oldFibonacci
		oldFibonacci = saveFibo
		res := &calculatorpb.FibonacciResponse{
			Result:     fibonacci,
			Iterations: i,
		}
		stream.Send(res)
	}
	return nil
}

func (*server) Average(stream calculatorpb.CalculatorService_AverageServer) error {
	fmt.Println("Start Average")
	var sum float32
	var numberOfNumbers float32
	for {
		time.Sleep(500 * time.Millisecond)
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading : %v", req)
			return err
		}
		nb := req.GetNumber()
		sum += nb
		numberOfNumbers++
		sendErr := stream.Send(&calculatorpb.AverageResponse{
			Result: sum / numberOfNumbers,
		})
		if sendErr != nil {
			log.Fatalf("Error sending : %v", sendErr)
		}
	}
}
func main() {
	fmt.Println("Server ready")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Error Listen : %v", err)
	}
	certFile := "../ssl/server.crt"
	keyFile := "../ssl/server.pem"
	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
	if sslErr != nil {
		log.Fatalf("Failed loading certificates: %v", sslErr)
		return
	}
	opts := grpc.Creds(creds)
	s := grpc.NewServer(opts)
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Echec du serveur: %v", err)
	}
}

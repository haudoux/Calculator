package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"microservices/calculatorpb"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {

	certFile := "../ssl/ca.crt"
	creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
	if sslErr != nil {
		log.Fatalf("Error ssl : %v", sslErr)
	}
	opts := grpc.WithTransportCredentials(creds)
	dial, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Connection error %v", err)
	}
	defer dial.Close()

	client := calculatorpb.NewCalculatorServiceClient(dial)

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		multiplication(client)
	}()
	go func() {
		defer wg.Done()
		add(client)
	}()
	go func() {
		defer wg.Done()
		average(client)
	}()
	go func() {
		defer wg.Done()
		fibonacci(client)
	}()
	wg.Wait()
}
func multiplication(client calculatorpb.CalculatorServiceClient) {
	fmt.Println("Start Multiplications")
	req := &calculatorpb.MultiplicationRequest{
		Numbers: &calculatorpb.SimpleCalc{
			FirstNb:  5,
			SecondNb: 2,
		},
	}
	res, err := client.Multiplication(context.Background(), req)
	if err != nil {
		log.Fatalf("Error Multiplication RPC: %v \n", err)
	}
	fmt.Printf("Response mulitplications : %v \n", res.Result)
}

func add(client calculatorpb.CalculatorServiceClient) {
	fmt.Println("Start Add")
	requests := []*calculatorpb.AddRequest{
		&calculatorpb.AddRequest{
			Number: 3,
		},
		&calculatorpb.AddRequest{
			Number: 5,
		},
		&calculatorpb.AddRequest{
			Number: 7,
		},
		&calculatorpb.AddRequest{
			Number: 1,
		},
		&calculatorpb.AddRequest{
			Number: 20,
		},
	}

	stream, err := client.Add(context.Background())
	if err != nil {
		log.Fatalf("Error Average RPC :  %v", err)
	}
	for _, req := range requests {
		stream.Send(req)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error response : %v", err)
	}
	fmt.Printf("Response Add %d \n", res.GetResult())
}

func average(client calculatorpb.CalculatorServiceClient) {
	fmt.Println("Start Average")
	requests := []*calculatorpb.AverageRequest{
		&calculatorpb.AverageRequest{
			Number: 3,
		},
		&calculatorpb.AverageRequest{
			Number: 5,
		},
		&calculatorpb.AverageRequest{
			Number: 7,
		},
		&calculatorpb.AverageRequest{
			Number: 1,
		},
		&calculatorpb.AverageRequest{
			Number: 20,
		},
	}

	stream, err := client.Average(context.Background())
	if err != nil {
		log.Fatalf("Error Average RPC :  %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for _, req := range requests {
			stream.Send(req)
			time.Sleep(250 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving : %v \n", err)
				close(waitc)
			}
			fmt.Printf("Result average : %v \n", req.GetResult())
		}
	}()
	<-waitc
}

func fibonacci(client calculatorpb.CalculatorServiceClient) {
	fmt.Println("Start Fibonnaci")
	req := &calculatorpb.FibonacciRequest{
		Iterations: 10,
	}
	resStream, err := client.Fibonacci(context.Background(), req)
	if err != nil {
		log.Fatalf("Error Fibonacci RPC : %v \n", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while receiving Fibonnaci : %v \n", err)
		}
		fmt.Printf("Response Fibionnaci\n Iterations %d Fibonacci %d \n", msg.GetIterations(), msg.GetResult())
	}
}

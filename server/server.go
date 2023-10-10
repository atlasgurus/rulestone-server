package server

import (
	"fmt"
	grpc2 "github.com/atlasgurus/rulestone-server/grpc"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func StartGrpcServer(connType, address string) {
	lis, err := net.Listen(connType, address)
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	grpc2.RegisterRulestoneServiceServer(grpcServer, NewGrpcRulestoneServer())

	// Graceful shutdown handling
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		grpcServer.GracefulStop()
		if connType == "unix" {
			os.Remove(address)
		}
	}()

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v\n", err)
		os.Exit(1)
	}
}

func StartRulestoneServer(connType, address string) {
	listener, err := net.Listen(connType, address)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// Graceful shutdown handling
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		listener.Close()
		if connType == "unix" {
			os.Remove(address)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Stop accepting when listener is closed
			if opErr, ok := err.(*net.OpError); ok && opErr.Op == "accept" {
				break
			}
			panic(err)
		}
		rulestoneServer := NewRulestoneServer(conn)
		go rulestoneServer.HandleConnection()
		fmt.Println("Connection closed", conn)
	}
}

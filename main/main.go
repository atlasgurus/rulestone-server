package main

import (
	"flag"
	"fmt"
	grpc2 "github.com/atlasgurus/rulestone-server/grpc"
	"github.com/atlasgurus/rulestone-server/server"
	"google.golang.org/grpc"
	"net"
	"os"
)

func main() {
	var rpcType string
	var port int
	flag.StringVar(&rpcType, "rpc", "native", "Type of RPC to use: native or grpc")
	flag.IntVar(&port, "port", 50051, "The server port")
	flag.Parse()

	if rpcType == "native" {
		fmt.Printf("Starting native server..")
		startRulestoneServer()
	} else if rpcType == "grpc" {
		fmt.Printf("Starting grpc server..")
		startGrpcServer(port)
	} else {
		fmt.Printf("Invalid RPC type argument: %s", rpcType)
	}
}

func startGrpcServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer()

	grpc2.RegisterRulestoneServiceServer(grpcServer, server.NewGrpcRulestoneServer())

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Failed to serve: %v\n", err)
		os.Exit(1)
	}
}
func startRulestoneServer() {
	if _, err := os.Stat(server.SocketPath); err == nil {
		os.Remove(server.SocketPath)
	}

	listener, err := net.Listen("unix", server.SocketPath)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		rulestoneServer := server.NewRulestoneServer(conn)
		go rulestoneServer.HandleConnection()
		fmt.Println("Connection closed", conn)
	}
}

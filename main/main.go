package main

import (
	"flag"
	"fmt"
	grpc2 "github.com/atlasgurus/rulestone-server/grpc"
	"github.com/atlasgurus/rulestone-server/server"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var rpcType, connType, unixPath string
	var port int

	flag.StringVar(&rpcType, "rpc", "native", "Type of RPC to use: native or grpc")
	flag.StringVar(&connType, "conn", "unix", "Type of connection to use: tcp or unix")
	flag.IntVar(&port, "port", 50051, "The server port (for tcp sockets)")
	flag.StringVar(&unixPath, "unixpath", server.SocketPath, "The file path for UNIX sockets")
	flag.Parse()

	address := getAddress(connType, port, unixPath)

	if connType == "unix" && fileExists(address) {
		os.Remove(address)
	}

	if rpcType == "native" {
		fmt.Printf("Starting native server using %s socket: %s\n", connType, address)
		startRulestoneServer(connType, address)
	} else if rpcType == "grpc" {
		fmt.Printf("Starting grpc server using %s socket: %s\n", connType, address)
		startGrpcServer(connType, address)
	} else {
		fmt.Printf("Invalid RPC type argument: %s", rpcType)
		os.Exit(1)
	}
}

func getAddress(connType string, port int, unixPath string) string {
	if connType == "tcp" {
		return fmt.Sprintf(":%d", port)
	} else {
		return unixPath
	}
}

func startGrpcServer(connType, address string) {
	lis, err := net.Listen(connType, address)
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	grpc2.RegisterRulestoneServiceServer(grpcServer, server.NewGrpcRulestoneServer())

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

func startRulestoneServer(connType, address string) {
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
		rulestoneServer := server.NewRulestoneServer(conn)
		go rulestoneServer.HandleConnection()
		fmt.Println("Connection closed", conn)
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

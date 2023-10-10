package main

import (
	"flag"
	"fmt"
	"github.com/atlasgurus/rulestone-server/server"
	"os"
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
		server.StartRulestoneServer(connType, address)
	} else if rpcType == "grpc" {
		fmt.Printf("Starting grpc server using %s socket: %s\n", connType, address)
		server.StartGrpcServer(connType, address)
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

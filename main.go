package main

import (
	pb "drcache/grpc/definitions"
	"drcache/src"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

var (
	allServers = map[string]struct{}{"localhost:50051": {}}
)

func main() {
	self := os.Args[1]
	lis, err := net.Listen("tcp", self)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	drcacheServer := src.NewServer(allServers, 3, self)
	grpcServer := grpc.NewServer()
	pb.RegisterDrcacheServer(grpcServer, drcacheServer)
	println("Server is started.")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

package main

import (
	dc "drcache/grpc"
	pb "drcache/grpc/definitions"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	self = "localhost:50051"
)

var (
	allServers = []string{"localhost:50051", "localhost:50052"}
)

func main() {
	lis, err := net.Listen("tcp", self)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	drcacheServer := dc.NewServer(allServers, 1024, self)
	grpcServer := grpc.NewServer()
	pb.RegisterDrcacheServer(grpcServer, drcacheServer)
	println("Server is started.")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

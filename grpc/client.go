package grpc

import (
	"context"
	pb "drcache/grpc/definitions"
	"google.golang.org/grpc"
	"log"
)

func AddItem(address string) pb.Reply {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

}

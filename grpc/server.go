package grpc

import (
	"context"
	pb "drcache/grpc/definitions"
	"log"
)

const (
	port = ":50051"
)

type server struct{}

//----------------------------------------------------------
// With consistent hashing check if key belogs to you, if so add to local cache. Otherwise send to other server with client
//----------------------------------------------------------
func (s *server) AddItem(ctx context.Context, in *pb.AddRequest) (*pb.Reply, error) {

	log.Printf("Received: %v", in.Item.Key)
	return &pb.Reply{Message: "ok"}, nil
}

package grpc

import (
	"context"
	"drcache/consistent_hashing"
	pb "drcache/grpc/definitions"
	"errors"
	lru "github.com/coocood/freecache"
	"log"
)

var cacheMissError = errors.New("Key does not exist.")

type Server struct {
	lru         *lru.Cache
	ch          *consistent_hashing.Ring
	serverList  []string
	selfAddress string
	client      *Client
}

// With consistent hashing check if key belongs to you, if so add to local cache. Otherwise send to other server with client
func (s *Server) Add(ctx context.Context, in *pb.AddRequest) (*pb.Reply, error) {
	key := in.Item.Key
	log.Printf("Received: %v", key)
	value := in.Item.Value
	expiration := in.Item.Expiration
	nodeAddress := s.ch.Get(key)
	if nodeAddress == s.selfAddress {
		err := s.lru.AddItem(key, value, expiration)
		return &pb.Reply{Message: "ok"}, err
	} else {
		return s.client.AddItem(nodeAddress, in)
	}
}

func (s *Server) CompareAndSwap(ctx context.Context, in *pb.CompareAndSwapRequest) (*pb.Reply, error) {
	log.Printf("Received: %v", in.Item.Key)
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Decrement(ctx context.Context, in *pb.DecrementRequest) (*pb.Reply, error) {
	log.Printf("Received: %v", in.Key)
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Increment(ctx context.Context, in *pb.IncrementRequest) (*pb.Reply, error) {
	key := in.Key
	delta := in.Delta
	log.Printf("Received: %v", in.Key)
	retval := s.lru.IncrementItem(key, delta)
	if !retval {
		return &pb.Reply{Message: "NOT OK!"}, cacheMissError
	}
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Replace(ctx context.Context, in *pb.ReplaceRequest) (*pb.Reply, error) {
	log.Printf("Received: %v", in.Item.Key)
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Set(ctx context.Context, in *pb.SetRequest) (*pb.Reply, error) {
	log.Printf("Received: %v", in.Item.Key)
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Touch(ctx context.Context, in *pb.TouchRequest) (*pb.Reply, error) {
	log.Printf("Received: %v", in.Key)
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Delete(ctx context.Context, in *pb.DeleteRequest) (*pb.Reply, error) {
	log.Printf("Received: %v", in.Key)
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) DeleteAll(ctx context.Context, in *pb.DeleteAllRequest) (*pb.Reply, error) {
	log.Printf("Received: %v", "deleteAll")
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.Reply, error) {
	nodeAddress := s.ch.Get(in.Key)
	if nodeAddress == s.selfAddress {
		i, ok := s.lru.GetItem(in.Key)
		if ok {
			return &pb.Reply{Message: "ok", Item: &pb.Item{Key: i.GetKey(), Value: i.GetValue(), Expiration: i.GetExpiration()}}, nil
		}
		return nil, nil
	} else {
		return s.client.getItem(nodeAddress, in)
	}
}

func (s *Server) AddServer(ctx context.Context, in *pb.AddServerRequest) (*pb.Reply, error) {
	panic("implement me")
}

func (s *Server) DropServer(ctx context.Context, in *pb.DropServerRequest) (*pb.Reply, error) {
	panic("implement me")
}

func (s *Server) CheckConnection(ctx context.Context, in *pb.CheckConnectionRequest) (*pb.Reply, error) {
	panic("implement me")
}

func NewServer(ipList []string, maxSize int, localAddress string) *Server {
	cache := lru.NewCache(maxSize)
	ch := consistent_hashing.NewRing(ipList)
	return &Server{lru: cache, ch: ch, serverList: ipList, selfAddress: localAddress, client: NewClient(ipList, localAddress)}
}

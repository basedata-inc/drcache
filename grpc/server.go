package grpc

import (
	"context"
	"drcache/consistent_hashing"
	pb "drcache/grpc/definitions"
	"errors"
	lru "github.com/coocood/freecache"
	"google.golang.org/grpc/status"
	"sync"

	utils "drcache"
	"log"
)

var cacheMissError = errors.New("Key does not exist.")

type Server struct {
	lru         *lru.Cache
	ch          *consistent_hashing.Ring
	serverList  map[string]struct{} // set of servers
	selfAddress string
	client      *Client
	sync.Mutex
}

/* With consistent hashing check if key belongs to you, if so add to local cache. Otherwise send to other server with client
   Adds if key does not exist already.
   If key exists, returns error
*/
func (s *Server) Add(ctx context.Context, in *pb.AddRequest) (*pb.Reply, error) {
	key := in.Item.Key
	log.Printf("Received: %v", key)
	value := in.Item.Value
	expiration := in.Item.Expiration
	nodeAddress := s.ch.Get(key)
	if nodeAddress == s.selfAddress {
		getval, _ := s.lru.Get([]byte(key))
		if getval != nil {
			return &pb.Reply{Message: "Key already exists."}, nil
		} else {
			err := s.lru.Set([]byte(key), value, int(expiration))
			return &pb.Reply{Message: "ok"}, err
		}
	} else {
		reply, err := s.client.AddItem(nodeAddress, in)
		if status.Code(err) == 14 { // Connection Error server is down
			s.dropAndReInit(nodeAddress)
		}
		return reply, err
	}
}

/* With consistent hashing check if key belongs to you, if so add to local cache. Otherwise send to other server with client
If entry does not exist, adds the entry.
If exists updates the entry's value.
*/
func (s *Server) Set(ctx context.Context, in *pb.SetRequest) (*pb.Reply, error) {
	key := in.Item.Key
	log.Printf("Received: %v", key)
	value := in.Item.Value
	expiration := in.Item.Expiration
	nodeAddress := s.ch.Get(key)
	if nodeAddress == s.selfAddress {
		err := s.lru.Set([]byte(key), value, int(expiration))
		return &pb.Reply{Message: "ok"}, err
	} else {
		reply, err := s.client.SetItem(nodeAddress, in)
		if status.Code(err) == 14 { // Connection Error server is down
			s.dropAndReInit(nodeAddress)
		}
		return reply, err
	}
}

/* With consistent hashing check if key belongs to you, if so add to local cache. Otherwise send to other server with client
If entry does not exist, return error.
If exists deletes the entry
*/
func (s *Server) Delete(ctx context.Context, in *pb.DeleteRequest) (*pb.Reply, error) {
	nodeAddress := s.ch.Get(in.Key)
	if nodeAddress == s.selfAddress {
		ret := s.lru.Del([]byte(in.Key))
		if ret {
			return &pb.Reply{Message: "ok"}, nil
		}
		return nil, errors.New("not found")

	} else {
		return s.client.DeleteItem(nodeAddress, in)
	}
}

/*
flushes all cache
*/
func (s *Server) DeleteAll(ctx context.Context, in *pb.DeleteAllRequest) (*pb.Reply, error) {
	s.lru.Clear()
	return &pb.Reply{Message: "ok"}, nil
}

func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.Reply, error) {
	nodeAddress := s.ch.Get(in.Key)
	if nodeAddress == s.selfAddress {
		val, exp, err := s.lru.GetWithExpiration([]byte(in.Key))
		if err == nil {
			return &pb.Reply{Message: "ok", Item: &pb.Item{Key: in.Key, Value: val, Expiration: exp}}, nil
		}
		return nil, err
	} else {
		reply, err := s.client.GetItem(nodeAddress, in)
		if status.Code(err) == 14 { // Connection Error server is down
			s.dropAndReInit(nodeAddress)
		}
		return reply, err
	}
}

func (s *Server) AddServer(ctx context.Context, in *pb.AddServerRequest) (*pb.Reply, error) {
	s.Lock()
	defer s.Unlock()
}

func (s *Server) DropServer(ctx context.Context, in *pb.DropServerRequest) (*pb.Reply, error) {
	s.dropAndReInit(in.Server)
	return nil, nil
}

func NewServer(ipList map[string]struct{}, maxSize int, localAddress string) *Server {
	cache := lru.NewCache(maxSize)
	ch := consistent_hashing.NewRing(ipList)
	return &Server{lru: cache, ch: ch, serverList: ipList, selfAddress: localAddress, client: NewClient(ipList, localAddress)}
}

func (s *Server) dropAndReInit(deadNode string) {
	s.Lock()
	defer s.Unlock()
	delete(s.serverList, deadNode)
	s.ch = consistent_hashing.NewRing(s.serverList)
	for address, _ := range s.serverList {
		if address != s.selfAddress {
			go s.client.DropServer(address)
		}
	}
}

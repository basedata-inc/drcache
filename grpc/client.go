package grpc

import (
	"context"
	pb "drcache/grpc/definitions"
	"google.golang.org/grpc"
	"log"
	"sync"
)

type Client struct {
	Clients map[string]pb.DrcacheClient
	sync.Mutex
}

var client *Client
var once sync.Once

func NewClient(ServerList []string, self string) *Client {

	once.Do(func() { // <-- atomic, does not allow repeating
		clients := make(map[string]pb.DrcacheClient)
		for _, address := range ServerList {
			if address == self {
				continue
			}
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect to %s: %v", address, err)
			}
			c := pb.NewDrcacheClient(conn)
			clients[address] = c
		}
		client = &Client{Clients: clients}
	})
	return client
}

func (c *Client) AddItem(address string, request pb.AddRequest) (*pb.Reply, error) {
	return c.Clients[address].Add(context.Background(), &request)
}

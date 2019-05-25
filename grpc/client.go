package grpc

import (
	"context"
	pb "drcache/grpc/definitions"
)

type Client struct {
	Clients map[string]pb.DrcacheClient
}

func (c *Client) AddItem(address string, request pb.AddRequest) (*pb.Reply, error) {
	return c.Clients[address].Add(context.Background(), &request)
}

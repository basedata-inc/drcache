package consistent_hashing

import (
	"encoding/json"
	"errors"
	"hash/crc32"
	"sort"
	"sync"
)

var ErrNodeNotFound = errors.New("node not found")

const numberOfReplicas = 10

//----------------------------------------------------------
// Node
//----------------------------------------------------------

type Node struct {
	Id        string
	ReplicaId int
	HashId    uint32
}

func newNode(id string, replicaId int) *Node {
	return &Node{
		Id:        id,
		ReplicaId: replicaId,
		HashId:    hashByKeyAndId(id, replicaId)}
}

type Nodes []*Node

//----------------------------------------------------------
// Ring
//----------------------------------------------------------

type Ring struct {
	Nodes Nodes
	sync.Mutex
}

func NewRing(serverList map[string]struct{}) *Ring {
	ring := &Ring{Nodes: Nodes{}}
	for address := range serverList {
		ring.AddNode(address)
	}
	return ring
}

func (r *Ring) AddNode(id string) {
	r.Lock()
	defer r.Unlock()

	for i := 0; i < numberOfReplicas; i++ {
		node := newNode(id, i)
		r.Nodes = append(r.Nodes, node)
	}
	sort.Sort(r.Nodes)
}

func (r *Ring) RemoveNode(id string) error {
	r.Lock()
	defer r.Unlock()
	for i := 0; i < numberOfReplicas; i++ {
		index := r.search(hashByKeyAndId(id, i))
		if index >= r.Nodes.Len() || r.Nodes[index].Id != id {
			return ErrNodeNotFound
		}
		r.Nodes = append(r.Nodes[:index], r.Nodes[index+1:]...)
	}
	return nil
}

func (r *Ring) Get(key string) string {
	i := r.search(hashByKey(key))
	if i >= r.Nodes.Len() {
		i = 0
	}
	return r.Nodes[i].Id
}

// Search for Node with smallest hash which is greater than hash of new key.
// (Node, which is just on the right)
func (r *Ring) search(id uint32) int {
	searchfn := func(i int) bool {
		return r.Nodes[i].HashId >= id
	}
	return sort.Search(r.Nodes.Len(), searchfn)
}

func (n Nodes) Len() int           { return len(n) }
func (n Nodes) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n Nodes) Less(i, j int) bool { return n[i].HashId < n[j].HashId }

//----------------------------------------------------------
// Helpers
//----------------------------------------------------------
type Identity struct {
	Key       string
	ReplicaId int
}

func hashByKeyAndId(key string, replicaId int) uint32 {
	bytes, _ := json.Marshal(&Identity{key, replicaId})
	return crc32.ChecksumIEEE(bytes)
}

func hashByKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

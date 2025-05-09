package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const defaultTimeout = time.Second

type AddReq struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
}

type AddResp struct {
	Type string `json:"type"`
}

type ReadResp struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type LocalReq struct {
	Type string `json:"type"`
}

type LocalResp struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type server struct {
	n          *maelstrom.Node
	kv         *maelstrom.KV
	localValue int
	mu         sync.Mutex
}

func (s *server) addHandler(msg maelstrom.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var req AddReq
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		log.Printf("node %s: failed to unmarshal add request: %v", s.n.ID(), err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.localValue += req.Delta

	if err := s.kv.Write(ctx, s.n.ID(), s.localValue); err != nil {
		log.Printf("node %s: failed to write local value: %v", s.n.ID(), err)
		return err
	}

	return s.n.Reply(msg, AddResp{Type: "add_ok"})
}

func (s *server) readHandler(msg maelstrom.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var wg sync.WaitGroup
	sum := 0
	mu := sync.Mutex{}

	for _, nodeID := range s.n.NodeIDs() {
		wg.Add(1)
		go func(nodeID string) {
			defer wg.Done()

			var value int
			if nodeID == s.n.ID() {
				s.mu.Lock()
				value = s.localValue
				s.mu.Unlock()
			} else {
				res, err := s.n.SyncRPC(ctx, nodeID, LocalReq{Type: "local"})
				if err != nil {
					log.Printf("node %s: failed to read from node %s: %v", s.n.ID(), nodeID, err)
					return
				}

				var resp LocalResp
				if err := json.Unmarshal(res.Body, &resp); err != nil {
					log.Printf("node %s: failed to unmarshal local response from node %s: %v", s.n.ID(), nodeID, err)
					return
				}
				value = resp.Value
			}

			mu.Lock()
			sum += value
			mu.Unlock()
		}(nodeID)
	}

	wg.Wait()

	return s.n.Reply(msg, ReadResp{
		Type:  "read_ok",
		Value: sum,
	})
}

func (s *server) localHandler(msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.n.Reply(msg, LocalResp{
		Type:  "local_ok",
		Value: s.localValue,
	})
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	s := &server{
		n:          n,
		kv:         kv,
		localValue: 0,
		mu:         sync.Mutex{},
	}

	n.Handle("add", s.addHandler)
	n.Handle("read", s.readHandler)
	n.Handle("local", s.localHandler)

	if err := n.Run(); err != nil {
		log.Fatalf("node run failed: %v", err)
	}
}

package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastReq struct {
	Type  string `json:"type"`
	Msg   int    `json:"message"`
	MsgID int    `json:"msg_id"`
}

type ReadResp struct {
	Type string `json:"type"`
	Msgs []int  `json:"messages"`
}

type server struct {
	node     *maelstrom.Node
	msgs     map[int]struct{}
	msgMutex sync.RWMutex
}

func initServer() *server {
	return &server{
		node: maelstrom.NewNode(),
		msgs: make(map[int]struct{}),
	}
}

func (s *server) handleBroadcast(msg maelstrom.Message) error {
	var req BroadcastReq
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	s.msgMutex.Lock()
	if _, exists := s.msgs[req.Msg]; exists {
		s.msgMutex.Unlock()
		return s.node.Reply(msg, map[string]string{"type": "broadcast_ok"})
	}
	s.msgs[req.Msg] = struct{}{}
	s.msgMutex.Unlock()
	for _, node := range s.node.NodeIDs() {
		if node == s.node.ID() || node == msg.Src {
			continue
		}
		go func(node string) {
			if err := s.node.Send(node, req); err != nil {
				log.Printf("Error propagating message to node %s: %v", node, err)
			}
		}(node)
	}
	return s.node.Reply(msg, map[string]string{"type": "broadcast_ok"})
}

func (s *server) handleRead(msg maelstrom.Message) error {
	s.msgMutex.RLock()
	msgs := make([]int, 0, len(s.msgs))
	for msg := range s.msgs {
		msgs = append(msgs, msg)
	}
	s.msgMutex.RUnlock()
	return s.node.Reply(msg, ReadResp{"read_ok", msgs})
}

func (s *server) handleTopology(msg maelstrom.Message) error {
	// ignore topology and broadcast to all nodes
	return s.node.Reply(msg, map[string]string{"type": "topology_ok"})
}

func main() {
	s := initServer()
	s.node.Handle("broadcast", s.handleBroadcast)
	s.node.Handle("read", s.handleRead)
	s.node.Handle("topology", s.handleTopology)
	if err := s.node.Run(); err != nil {
		log.Fatal(err)
	}
}

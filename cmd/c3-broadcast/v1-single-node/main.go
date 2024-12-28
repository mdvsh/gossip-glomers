package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastReq struct {
	Type string `json:"type"`
	Msg  int    `json:"message"`
}

type ReadResp struct {
	Type string `json:"type"`
	Msgs []int  `json:"messages"`
}

type TopologyReq struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

type server struct {
	node     *maelstrom.Node
	msgs     []int
	msgMutex sync.RWMutex
	top      map[string][]string
	topMutex sync.RWMutex
}

func initServer() *server {
	return &server{
		node: maelstrom.NewNode(),
		msgs: make([]int, 0),
		top:  make(map[string][]string),
	}
}

func (s *server) handleBroadcast(msg maelstrom.Message) error {
	var req BroadcastReq
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	s.msgMutex.Lock()
	s.msgs = append(s.msgs, req.Msg)
	s.msgMutex.Unlock()
	return s.node.Reply(msg, map[string]string{"type": "broadcast_ok"})

}

func (s *server) handleRead(msg maelstrom.Message) error {
	s.msgMutex.Lock()
	msgs := make([]int, len(s.msgs))
	copy(msgs, s.msgs)
	s.msgMutex.Unlock()
	return s.node.Reply(msg, ReadResp{Type: "read_ok", Msgs: msgs})
}

func (s *server) handleTopology(msg maelstrom.Message) error {
	var req TopologyReq
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}
	s.topMutex.Lock()
	s.top = req.Topology
	s.topMutex.Unlock()
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

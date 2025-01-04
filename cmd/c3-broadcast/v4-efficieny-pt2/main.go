package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastReq struct {
	Type  string `json:"type"`
	Msg   int    `json:"message"`
	MsgID int    `json:"msg_id"`
}

type SyncMessage struct {
	Type string `json:"type"`
	Msgs []int  `json:"messages"`
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
	s := &server{
		node: maelstrom.NewNode(),
		msgs: make(map[int]struct{}),
	}
	go s.startPeriodicSync()
	return s
}

func (s *server) startPeriodicSync() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		s.msgMutex.RLock()
		msgs := make([]int, 0, len(s.msgs))
		for msg := range s.msgs {
			msgs = append(msgs, msg)
		}
		s.msgMutex.RUnlock()

		syncMsg := SyncMessage{"sync", msgs}

		for _, node := range s.node.NodeIDs() {
			if node == s.node.ID() {
				continue
			}
			go func(node string) {
				if err := s.node.Send(node, syncMsg); err != nil {
					log.Printf("Error syncing with node %s: %v", node, err)
				}
			}(node)
		}
	}
}

func (s *server) handleSync(msg maelstrom.Message) error {
	var req SyncMessage
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}

	s.msgMutex.Lock()
	for _, m := range req.Msgs {
		s.msgs[m] = struct{}{}
	}
	s.msgMutex.Unlock()

	return nil
}

func (s *server) handleBroadcast(msg maelstrom.Message) error {
	var req BroadcastReq
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		return err
	}

	s.msgMutex.Lock()
	s.msgs[req.Msg] = struct{}{}
	s.msgMutex.Unlock()

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
	return s.node.Reply(msg, map[string]string{"type": "topology_ok"})
}

func main() {
	s := initServer()
	s.node.Handle("broadcast", s.handleBroadcast)
	s.node.Handle("read", s.handleRead)
	s.node.Handle("topology", s.handleTopology)
	s.node.Handle("sync", s.handleSync)

	if err := s.node.Run(); err != nil {
		log.Fatal(err)
	}
}

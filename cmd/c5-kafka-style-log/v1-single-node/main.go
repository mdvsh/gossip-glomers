package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type KafkaServer struct {
	node *maelstrom.Node
	mu   sync.RWMutex
	logs map[string]*Log
}

type Log struct {
	messages        map[int]any
	nextOffset      int
	committedOffset int
}

func NewKafkaServer(node *maelstrom.Node) *KafkaServer {
	return &KafkaServer{
		node: node,
		logs: make(map[string]*Log),
	}
}

func (s *KafkaServer) getOrCreateLog(key string) *Log {
	log, exists := s.logs[key]
	if !exists {
		log = &Log{
			messages:        make(map[int]any),
			nextOffset:      0,
			committedOffset: -1,
		}
		s.logs[key] = log
	}
	return log
}

func (s *KafkaServer) sendHandler(msg maelstrom.Message) error {
	var body struct {
		Type string `json:"type"`
		Key  string `json:"key"`
		Msg  any    `json:"msg"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	log := s.getOrCreateLog(body.Key)
	offset := log.nextOffset
	log.messages[offset] = body.Msg
	log.nextOffset++
	s.mu.Unlock()

	return s.node.Reply(msg, map[string]any{
		"type":   "send_ok",
		"offset": offset,
	})
}

func (s *KafkaServer) pollHandler(msg maelstrom.Message) error {
	var body struct {
		Type    string         `json:"type"`
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.RLock()
	response := make(map[string][][2]any)

	for key, startOffset := range body.Offsets {
		log, exists := s.logs[key]
		if !exists {
			continue
		}

		var messages [][2]any
		currentOffset := startOffset

		for {
			msg, exists := log.messages[currentOffset]
			if !exists {
				break
			}
			messages = append(messages, [2]any{currentOffset, msg})
			currentOffset++
		}

		if len(messages) > 0 {
			response[key] = messages
		}
	}
	s.mu.RUnlock()

	return s.node.Reply(msg, map[string]any{
		"type": "poll_ok",
		"msgs": response,
	})
}

func (s *KafkaServer) commitOffsetsHandler(msg maelstrom.Message) error {
	var body struct {
		Type    string         `json:"type"`
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	for key, offset := range body.Offsets {
		log, exists := s.logs[key]
		if exists {
			log.committedOffset = offset
		}
	}
	s.mu.Unlock()

	return s.node.Reply(msg, map[string]any{
		"type": "commit_offsets_ok",
	})
}

func (s *KafkaServer) listCommittedOffsetsHandler(msg maelstrom.Message) error {
	var body struct {
		Type string   `json:"type"`
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.RLock()
	offsets := make(map[string]int)
	for _, key := range body.Keys {
		if log, exists := s.logs[key]; exists && log.committedOffset >= 0 {
			offsets[key] = log.committedOffset
		}
	}
	s.mu.RUnlock()

	return s.node.Reply(msg, map[string]any{
		"type":    "list_committed_offsets_ok",
		"offsets": offsets,
	})
}

func main() {
	node := maelstrom.NewNode()
	server := NewKafkaServer(node)

	node.Handle("send", server.sendHandler)
	node.Handle("poll", server.pollHandler)
	node.Handle("commit_offsets", server.commitOffsetsHandler)
	node.Handle("list_committed_offsets", server.listCommittedOffsetsHandler)

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}

package main

/*

Community
Github
Twitter
Support forum
Challenge #5a: Single-Node Kafka-Style Log

In this challenge, you’ll need to implement a replicated log service similar to Kafka. Replicated logs are often used as a message bus or an event stream.

This challenge is broken up in multiple sections so that you can build out your system incrementally. First, we’ll start out with a single-node log system and then we’ll distribute it in later challenges.
Specification

Your nodes will need to store an append-only log in order to handle the "kafka" workload. Each log is identified by a string key (e.g. "k1") and these logs contain a series of messages which are identified by an integer offset. These offsets can be sparse in that not every offset must contain a message.

Maelstrom will check to make sure several anomalies do not occur:

    Lost writes: for example, a client sees offset 10 but not offset 5.
    Monotonic increasing offsets: an offset for a log should always be increasing.

There are no recency requirements so acknowledged send messages do not need to return in poll messages immediately.
RPC: send

This message requests that a "msg" value be appended to a log identified by "key". Your node will receive a request message body that looks like this:

{
  "type": "send",
  "key": "k1",
  "msg": 123
}

In response, it should send an acknowledge with a send_ok message that contains the unique offset for the message in the log:

{
  "type": "send_ok",
  "offset": 1000
}

RPC: poll

This message requests that a node return messages from a set of logs starting from the given offset in each log. Your node will receive a request message body that looks like this:

{
  "type": "poll",
  "offsets": {
    "k1": 1000,
    "k2": 2000
  }
}

In response, it should return a poll_ok message with messages starting from the given offset for each log. Your server can choose to return as many messages for each log as it chooses:

{
  "type": "poll_ok",
  "msgs": {
    "k1": [[1000, 9], [1001, 5], [1002, 15]],
    "k2": [[2000, 7], [2001, 2]]
  }
}

RPC: commit_offsets

This message informs the node that messages have been successfully processed up to and including the given offset. Your node will receive a request message body that looks like this:

{
  "type": "commit_offsets",
  "offsets": {
    "k1": 1000,
    "k2": 2000
  }
}

In this example, the messages have been processed up to and including offset 1000 for log k1 and all messages up to and including offset 2000 for k2.

In response, your node should return a commit_offsets_ok message body to acknowledge the request:

{
  "type": "commit_offsets_ok"
}

RPC: list_committed_offsets

This message returns a map of committed offsets for a given set of logs. Clients use this to figure out where to start consuming from in a given log.

Your node will receive a request message body that looks like this:

{
  "type": "list_committed_offsets",
  "keys": ["k1", "k2"]
}

In response, your node should return a list_committed_offsets_ok message body containing a map of offsets for each requested key. Keys that do not exist on the node can be omitted.

{
  "type": "list_committed_offsets_ok",
  "offsets": {
    "k1": 1000,
    "k2": 2000
  }
}


*/

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

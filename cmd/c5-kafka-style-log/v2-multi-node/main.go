package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type KafkaServer struct {
	node *maelstrom.Node
	kv   *maelstrom.KV
}

func NewKafkaServer(node *maelstrom.Node) *KafkaServer {
	return &KafkaServer{
		node: node,
		kv:   maelstrom.NewLinKV(node),
	}
}

func msgKey(topic string, offset int) string {
	return fmt.Sprintf("msg:%s:%d", topic, offset)
}

func offsetKey(topic string) string {
	return fmt.Sprintf("next:%s", topic)
}

func commitKey(topic string) string {
	return fmt.Sprintf("commit:%s", topic)
}

func toInt(v any) (int, error) {
	switch x := v.(type) {
	case float64:
		return int(x), nil
	case int:
		return x, nil
	case int64:
		return int(x), nil
	default:
		return 0, fmt.Errorf("unexpected type %T for integer value", v)
	}
}

// atomically allocates and returns the next available offset
func (s *KafkaServer) reserveOffset(ctx context.Context, topic string) (int, error) {
	key := offsetKey(topic)
	
	for {
		// Try to read current offset
		val, err := s.kv.Read(ctx, key)
		if err != nil {
			// Initialize if this topic is new
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				err = s.kv.CompareAndSwap(ctx, key, 0, 1, true)
				if err == nil {
					return 0, nil // Successfully initialized with offset 0
				}
				if maelstrom.ErrorCode(err) == maelstrom.PreconditionFailed {
					continue // Someone else initialized it, retry
				}
				return 0, err // Unexpected error
			}
			return 0, err // Propagate other errors
		}

		// Convert to int (handles JSON number type inconsistencies)
		current, err := toInt(val)
		if err != nil {
			return 0, err
		}

		// Atomically increment
		err = s.kv.CompareAndSwap(ctx, key, current, current+1, false)
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.PreconditionFailed {
				continue // Someone else updated it, retry
			}
			return 0, err
		}
		
		return current, nil // Return current offset (not the incremented value)
	}
}

// processes send requests by reserving an offset and storing the message
func (s *KafkaServer) sendHandler(msg maelstrom.Message) error {
	var body struct {
		Type string `json:"type"`
		Key  string `json:"key"`
		Msg  any    `json:"msg"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	
	// Reserve next offset
	offset, err := s.reserveOffset(ctx, body.Key)
	if err != nil {
		return err
	}

	// Store message
	if err := s.kv.Write(ctx, msgKey(body.Key, offset), body.Msg); err != nil {
		return err
	}

	return s.node.Reply(msg, map[string]any{
		"type":   "send_ok",
		"offset": offset,
	})
}

// retrieves messages from requested offsets
func (s *KafkaServer) pollHandler(msg maelstrom.Message) error {
	var body struct {
		Type    string         `json:"type"`
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	result := make(map[string][][2]any)

	for topic, startOffset := range body.Offsets {
		// Get current next offset to determine range
		val, err := s.kv.Read(ctx, offsetKey(topic))
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				continue // Topic doesn't exist yet
			}
			return err
		}
		
		endOffset, err := toInt(val)
		if err != nil {
			return err
		}

		// Collect all available messages
		var messages [][2]any
		for offset := startOffset; offset < endOffset; offset++ {
			val, err := s.kv.Read(ctx, msgKey(topic, offset))
			if err != nil {
				if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
					continue // Skip gaps in the log
				}
				return err
			}
			messages = append(messages, [2]any{offset, val})
		}

		if len(messages) > 0 {
			result[topic] = messages
		}
	}

	return s.node.Reply(msg, map[string]any{
		"type": "poll_ok",
		"msgs": result,
	})
}

// records consumer committed offsets
func (s *KafkaServer) commitOffsetsHandler(msg maelstrom.Message) error {
	var body struct {
		Type    string         `json:"type"`
		Offsets map[string]int `json:"offsets"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	
	// Write each committed offset
	for topic, offset := range body.Offsets {
		if err := s.kv.Write(ctx, commitKey(topic), offset); err != nil {
			return err
		}
	}

	return s.node.Reply(msg, map[string]any{
		"type": "commit_offsets_ok",
	})
}

// retrieves previously committed offsets
func (s *KafkaServer) listCommittedOffsetsHandler(msg maelstrom.Message) error {
	var body struct {
		Type string   `json:"type"`
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	offsets := make(map[string]int)
	
	// Read committed offsets for requested topics
	for _, topic := range body.Keys {
		val, err := s.kv.Read(ctx, commitKey(topic))
		if err != nil {
			if maelstrom.ErrorCode(err) == maelstrom.KeyDoesNotExist {
				continue // No committed offset for this topic
			}
			return err
		}
		
		offset, err := toInt(val)
		if err != nil {
			return err
		}
		
		offsets[topic] = offset
	}

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
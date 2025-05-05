package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type txnOperation []interface{}

type server struct {
	n    *maelstrom.Node
	kv   map[int]int
	lock sync.RWMutex
}

func main() {
	s := &server{
		n:  maelstrom.NewNode(),
		kv: make(map[int]int),
	}

	s.n.Handle("txn", s.handleTxn)

	if err := s.n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleTxn(msg maelstrom.Message) error {
	var body map[string]interface{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	txnOps := body["txn"].([]interface{})
	resultOps := make([]txnOperation, len(txnOps))

	s.lock.Lock()
	defer s.lock.Unlock()

	// Process each operation in the transaction
	for i, op := range txnOps {
		opArray := op.([]interface{})
		opType := opArray[0].(string)
		key := int(opArray[1].(float64))

		if opType == "r" {
			// Read operation
			value, exists := s.kv[key]
			var readValue interface{} = nil
			if exists {
				readValue = value
			}
			resultOps[i] = txnOperation{opType, key, readValue}
		} else if opType == "w" {
			// Write operation
			value := int(opArray[2].(float64))
			s.kv[key] = value
			resultOps[i] = txnOperation{opType, key, value}
		}
	}

	return s.n.Reply(msg, map[string]interface{}{
		"type": "txn_ok",
		"txn":  resultOps,
	})
}
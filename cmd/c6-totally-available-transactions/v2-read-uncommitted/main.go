package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type txnOperation []interface{}

type server struct {
	n        *maelstrom.Node
	kv       map[int]int
	lock     sync.RWMutex
	nodeIDs  []string
	topology map[string][]string
}

func main() {
	s := &server{
		n:        maelstrom.NewNode(),
		kv:       make(map[int]int),
		topology: make(map[string][]string),
	}

	s.n.Handle("txn", s.handleTxn)
	s.n.Handle("replicate", s.handleReplicate)
	s.n.Handle("topology", s.handleTopology)

	if err := s.n.Run(); err != nil {
		log.Fatal(err)
	}
}

func (s *server) handleTopology(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	topology := body["topology"].(map[string]any)
	for node, neighbors := range topology {
		s.topology[node] = make([]string, 0)
		for _, neighbor := range neighbors.([]any) {
			s.topology[node] = append(s.topology[node], neighbor.(string))
		}
	}

	// Store all node IDs for broadcasting
	if len(s.nodeIDs) == 0 {
		for nodeID := range s.topology {
			s.nodeIDs = append(s.nodeIDs, nodeID)
		}
	}

	return s.n.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
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
	writeOps := make([][]interface{}, 0)
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
			writeOps = append(writeOps, opArray)
		}
	}

	// Replicate writes to other nodes (async, don't wait for responses)
	if len(writeOps) > 0 {
		go s.broadcastWrites(writeOps)
	}

	return s.n.Reply(msg, map[string]interface{}{
		"type": "txn_ok",
		"txn":  resultOps,
	})
}

func (s *server) handleReplicate(msg maelstrom.Message) error {
	var body map[string]interface{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	writeOps := body["ops"].([]interface{})

	s.lock.Lock()
	defer s.lock.Unlock()

	// Apply the replicated writes locally
	for _, op := range writeOps {
		opArray := op.([]interface{})
		key := int(opArray[1].(float64))
		value := int(opArray[2].(float64))
		s.kv[key] = value
	}

	return s.n.Reply(msg, map[string]interface{}{
		"type": "replicate_ok",
	})
}

func (s *server) broadcastWrites(writeOps [][]interface{}) {
	for _, nodeID := range s.nodeIDs {
		if nodeID == s.n.ID() {
			continue // Skip ourselves
		}

		// Broadcast the write operations to other nodes
		err := s.n.Send(nodeID, map[string]interface{}{
			"type": "replicate",
			"ops":  writeOps,
		})

		if err != nil {
			// Log the error but continue - this is a "totally available" system,
			// so we proceed even if some replications fail
			log.Printf("Error replicating to %s: %v", nodeID, err)
		}
	}
}
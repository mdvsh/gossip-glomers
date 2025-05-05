package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type txnOperation []interface{}
type txnID string

// Versioned value to support read committed isolation
type versionedValue struct {
	value     int
	committed bool
}

// Transaction state tracking
type transaction struct {
	id       txnID
	writes   map[int]int // key -> value
	readKeys map[int]bool
}

type server struct {
	n           *maelstrom.Node
	kvStore     map[int][]versionedValue // key -> list of versioned values
	txns        map[txnID]*transaction   // active transactions
	lock        sync.RWMutex
	nodeIDs     []string
	topology    map[string][]string
	idGenerator int // for generating transaction IDs
}

func main() {
	s := &server{
		n:           maelstrom.NewNode(),
		kvStore:     make(map[int][]versionedValue),
		txns:        make(map[txnID]*transaction),
		topology:    make(map[string][]string),
		idGenerator: 0,
	}

	s.n.Handle("txn", s.handleTxn)
	s.n.Handle("commit", s.handleCommit)
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

func (s *server) generateTxnID() txnID {
	s.idGenerator++
	return txnID(s.n.ID() + "-" + string(rune(s.idGenerator)))
}

func (s *server) handleTxn(msg maelstrom.Message) error {
	var body map[string]interface{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	txnOps := body["txn"].([]interface{})
	resultOps := make([]txnOperation, len(txnOps))

	// Create a new transaction
	txID := s.generateTxnID()
	txn := &transaction{
		id:       txID,
		writes:   make(map[int]int),
		readKeys: make(map[int]bool),
	}

	s.lock.Lock()
	s.txns[txID] = txn
	s.lock.Unlock()

	// Process each operation in the transaction
	for i, op := range txnOps {
		opArray := op.([]interface{})
		opType := opArray[0].(string)
		key := int(opArray[1].(float64))

		if opType == "r" {
			// Read operation - only see committed values
			value, exists := s.getLatestCommittedValue(key)
			var readValue interface{} = nil
			if exists {
				readValue = value
			}
			resultOps[i] = txnOperation{opType, key, readValue}

			// Track that this transaction read this key
			txn.readKeys[key] = true
		} else if opType == "w" {
			// Write operation - buffer in transaction for now
			value := int(opArray[2].(float64))
			txn.writes[key] = value
			resultOps[i] = txnOperation{opType, key, value}
		}
	}

	// No explicit transaction abort/commit in this implementation
	// We'll immediately commit the transaction
	err := s.commitTransaction(txID)
	if err != nil {
		return s.abortTransaction(msg, txID)
	}

	// We've committed the transaction, so we can reply
	return s.n.Reply(msg, map[string]interface{}{
		"type": "txn_ok",
		"txn":  resultOps,
	})
}

func (s *server) handleCommit(msg maelstrom.Message) error {
	var body map[string]interface{}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// Extract transaction data
	writeData := body["writes"].([]interface{})

	s.lock.Lock()
	defer s.lock.Unlock()

	// Apply all writes as committed values
	for _, writeOp := range writeData {
		writeOpMap := writeOp.(map[string]interface{})
		key := int(writeOpMap["key"].(float64))
		value := int(writeOpMap["value"].(float64))

		// Add as a committed version
		if _, exists := s.kvStore[key]; !exists {
			s.kvStore[key] = make([]versionedValue, 0)
		}
		s.kvStore[key] = append(s.kvStore[key], versionedValue{
			value:     value,
			committed: true,
		})
	}

	return s.n.Reply(msg, map[string]interface{}{
		"type": "commit_ok",
	})
}

func (s *server) commitTransaction(id txnID) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	txn, exists := s.txns[id]
	if !exists {
		return nil // Transaction already gone (shouldn't happen)
	}

	// Apply all writes atomically to the local store
	for key, value := range txn.writes {
		if _, exists := s.kvStore[key]; !exists {
			s.kvStore[key] = make([]versionedValue, 0)
		}
		s.kvStore[key] = append(s.kvStore[key], versionedValue{
			value:     value,
			committed: true,
		})
	}

	// Replicate the committed transaction to other nodes asynchronously
	if len(txn.writes) > 0 {
		go s.replicateCommittedTransaction(txn)
	}

	// Remove the transaction from our tracking
	delete(s.txns, id)
	return nil
}

func (s *server) abortTransaction(msg maelstrom.Message, id txnID) error {
	s.lock.Lock()
	delete(s.txns, id) // Remove the transaction
	s.lock.Unlock()

	// Reply with a transaction conflict error
	return s.n.Reply(
		msg,
		map[string]any{
			"type": "error",
			"code": maelstrom.TxnConflict,
			"text": "txn abort",
		},
	)
}

func (s *server) getLatestCommittedValue(key int) (int, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	versions, exists := s.kvStore[key]
	if !exists || len(versions) == 0 {
		return 0, false
	}

	// Find the latest committed version
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].committed {
			return versions[i].value, true
		}
	}

	return 0, false
}

func (s *server) replicateCommittedTransaction(txn *transaction) {
	// Convert transaction writes to a format for replication
	writeData := make([]map[string]interface{}, 0, len(txn.writes))
	for key, value := range txn.writes {
		writeData = append(writeData, map[string]interface{}{
			"key":   key,
			"value": value,
		})
	}

	// Broadcast to all other nodes
	for _, nodeID := range s.nodeIDs {
		if nodeID == s.n.ID() {
			continue // Skip ourselves
		}

		// Send the committed writes
		err := s.n.Send(nodeID, map[string]interface{}{
			"type":   "commit",
			"writes": writeData,
		})

		if err != nil {
			// Log the error but continue - this is a "totally available" system,
			// so we proceed even if some replications fail
			log.Printf("Error replicating commit to %s: %v", nodeID, err)
		}
	}
}

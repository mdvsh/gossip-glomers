package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type server struct {
    n        *maelstrom.Node
    nodeID   string
    randLock sync.Mutex
}

// generateID creates a unique ID by combining:
// - current nano timestamp
// - node ID (already unique per node)
// - random bytes (for collision prevention)
func (s *server) generateID() string {
    now := time.Now().UnixNano()
    
    s.randLock.Lock()
    b := make([]byte, 6)
    _, err := rand.Read(b)
    s.randLock.Unlock()
    if err != nil {
        return fmt.Sprintf("%d-%s", now, s.nodeID)
    }

    return fmt.Sprintf("%d-%s-%s", now, s.nodeID, base64.URLEncoding.EncodeToString(b))
}

func main() {
    n := maelstrom.NewNode()
    s := &server{
        n:      n,
        nodeID: n.ID(),
    }

    n.Handle("generate", func(msg maelstrom.Message) error {
        uid := s.generateID()

        return n.Reply(msg, map[string]any{
            "type": "generate_ok",
            "id":   uid,
        })
    })

    if err := n.Run(); err != nil {
        log.Fatal(err)
    }
}
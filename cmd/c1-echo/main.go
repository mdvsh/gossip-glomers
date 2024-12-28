package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type EchoRequest struct {
    Type   string      `json:"type"`
    MsgID  int64       `json:"msg_id"`
    Echo   interface{} `json:"echo"`
}

type EchoResponse struct {
    Type      string      `json:"type"`
    MsgID     int64       `json:"msg_id,omitempty"`
    InReplyTo int64       `json:"in_reply_to,omitempty"`
    Echo      interface{} `json:"echo"`
}

func main() {
    n := maelstrom.NewNode()

    n.Handle("echo", func(msg maelstrom.Message) error {
        var req EchoRequest
        if err := json.Unmarshal(msg.Body, &req); err != nil {
            return err
        }

        resp := EchoResponse{
            Type: "echo_ok",
            Echo: req.Echo,
        }

        return n.Reply(msg, resp)
    })

    if err := n.Run(); err != nil {
        log.Fatal(err)
    }
}
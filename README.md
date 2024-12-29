## glomers

 Wrapped up distributed systems (eecs491) in college and loved it. There, I built a replicated, sharded key-value store with Paxos, and now am diving deeper into dist-sys patterns through [Fly.io's](https://fly.io/dist-sys/) challenges _(gossip glomers)_.

### What's Here
Working through a series of distributed systems challenges, each one building on the last:

1. [**Echo**](/cmd/c1-echo/README.md) - Getting our feet wet with the framework _(done)_
2. [**Unique ID Generation**](/cmd/c2-uid/README.md) - Making IDs that are actually unique across nodes _(done)_
3. **Broadcast** - Making nodes talk to each other reliably
   1. (3a) [Single-Node Broadcast](/cmd/c3-broadcast/v1-single-node/README.md) _(done)_
   2. (3b) [Multi-Node Broadcast](/cmd/c3-broadcast/v2-multi-node/README.md) _(done)_
4. **Grow-Only Counter** - Building a CRDT (fancy distributed counter)
5. **Kafka-Style Log** - Implementing a distributed log with some Kafka-like properties
6. **Transaction System** - Finale, handling distributed transactions

Each challenge has its own write-up explaining the key ideas and tradeoffs.

---

### Project Structure
```
glomers/
├── cmd/                    # Each challenge's implementation
│   ├── echo/              # Challenge 1 - Echo
│   ├── uid/               # Challenge 2 - Unique IDs
│   └── ...             
├── internal/              # Shared code 
│   └── transport/      
```

### Setup

You'll need:
- Go 1.20+
- Maelstrom 0.2.3 (the testing framework)
  - Needs Java (OpenJDK)
    - Download from Maelstrom's releases page

### Notes & Learning
Coming from implementing Paxos, it's cool to see other approaches to distributed systems problems. Each challenge tackles different aspects:
- How to handle node failures
- Ways to achieve consensus (or avoid needing it)
- Dealing with network partitions
- Making things eventually consistent

Will keep updating this as I work through more challenges.
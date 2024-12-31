## glomers

 Wrapped up distributed systems (eecs491) in college and loved it. There, I built a replicated, sharded key-value store with Paxos, and now am diving deeper into dist-sys patterns through [Fly.io's](https://fly.io/dist-sys/) challenges.
 
### What's Here
Working through a series of dist-sys challenges, each one building on the last:

1. [**Echo**](/cmd/c1-echo/README.md) - Getting our feet wet with the framework _(done)_
2. [**Unique ID Generation**](/cmd/c2-uid/README.md) - Making IDs that are actually unique across nodes _(done)_
3. **Broadcast** - Making nodes talk to each other reliably
   1. (3a) [Single-Node Broadcast](/cmd/c3-broadcast/v1-single-node/README.md) _(done)_
   2. (3b) [Multi-Node Broadcast](/cmd/c3-broadcast/v2-multi-node/README.md) _(done)_
   3. (3c) [Fault-Tolerant Broadcast](/cmd/c3-broadcast/v3-fault-tolerant/README.md) _(done)_
4. **Grow-Only Counter** - Building a CRDT (fancy distributed counter)
5. **Kafka-Style Log** - Implementing a distributed log with some Kafka-like properties
6. **Transaction System** - Finale, handling distributed transactions

Each challenge has its own write-up explaining the key ideas and tradeoffs.

---

### Project Structure
```
glomers/
├── cmd/                    # Challenge implementations
│   ├── c1-echo/             
│   ├── c2-uid/               
|   ├── c3-broadcast/
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

Execute `./run.sh` in a challenge directory to run the tests.
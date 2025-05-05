  ## glomers

  Wrapped up distributed systems (eecs491) in college and loved it. There, I built a replicated, sharded key-value store with Paxos, and now am diving deeper into dist-sys patterns through [Fly.io's](https://fly.io/dist-sys/)
  challenges.

  ### What's Here
  a series of dist-sys challenges, each one building on the last:

  1. [**Echo**](/cmd/c1-echo/README.md) - Getting our feet wet with the framework _(done)_
  2. [**Unique ID Generation**](/cmd/c2-uid/README.md) - Making IDs that are actually unique across nodes _(done)_
  3. **Broadcast** - Making nodes talk to each other reliably
     1. (3a) [Single-Node Broadcast](/cmd/c3-broadcast/v1-single-node/README.md) _(done)_
     2. (3b) [Multi-Node Broadcast](/cmd/c3-broadcast/v2-multi-node/README.md) _(done)_
     3. (3c) [Fault-Tolerant Broadcast](/cmd/c3-broadcast/v3-fault-tolerant/README.md) _(done)_
     4. (3d) Efficient Broadcast
        1. [Part I](/cmd/c3-broadcast/v4-efficieny-pt1/README.md) _(done)_
        2. [Part II](/cmd/c3-broadcast/v4-efficieny-pt2/README.md) _(done)_
  4. [**Grow-Only Counter**](/cmd/c4-grow-only-ctr/README.md) - Building a CRDT (fancy distributed counter) _(done)_
  5. **Kafka-Style Log** - Implementing a distributed log with some Kafka-like properties
     1. (5a) [Single-Node Kafka-Style Log](/cmd/c5-kafka-style-log/v1-single-node/README.md) _(done)_
     2. (5b) [Multi-Node Kafka-Style Log](/cmd/c5-kafka-style-log/v2-multi-node/README.md) _(done)_
     3. (5c) [Efficient Multi-Node Kafka-Style Log](/cmd/c5-kafka-style-log/v3-multi-node-efficient/README.md) _(done)_
  6. **Transaction System** - Building increasingly consistent distributed transactions
     1. (6a) [Single-Node Transactions](/cmd/c6-totally-available-transactions/v1-single-node/README.md) _(done)_
     2. (6b) [Read Uncommitted Transactions](/cmd/c6-totally-available-transactions/v2-read-uncommited/README.md) _(done)_
     3. (6c) [Read Committed Transactions](/cmd/c6-totally-available-transactions/v3-read-committed/README.md) _(done)_

  Each challenge has its own write-up explaining the key ideas and tradeoffs.

  ### Final Thoughts
  Finished all challenges! Started with simple echo servers, worked through message passing patterns, built CRDTs, replicated logs, and wrapped up with increasingly sophisticated distributed transaction systems - all while handling
  partitions and maintaining high availability. Still no silver bullets in distributed systems, but now I have a solid arsenal of patterns.

  ---

  ### Setup

  You'll need:
  - Go 1.20+
  - Maelstrom 0.2.3 (the testing framework)
    - Needs Java (OpenJDK)
      - Download from Maelstrom's releases page

  Execute `./run.sh` in a challenge directory to run the tests.
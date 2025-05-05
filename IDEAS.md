# Gossip Glomers - my impl overview

For detailed explanations, check out the README files in each challenge directory : )

## Challenge #1: Echo
**Quick Approach**: Basic message handler with JSON parsing and proper request/response correlation.

## Challenge #2: Unique ID Generation
**Core Idea**: Combined timestamps + node ID + random bytes for coordination-free unique IDs.

## Challenge #3: Broadcast
**Evolution of Approaches**:
- **Single-Node**: Thread-safe message storage with basic handlers
- **Multi-Node**: All-to-all broadcasting (simple but effective)
- **Fault-Tolerant**: Leveraged the same approach (natural resilience)
- **Efficient v1**: Two-level tree topology (root + children) reducing message count to ~23
- **Efficient v2**: Added message batching, dropping message count to ~1.92

## Challenge #4: Grow-Only Counter (CRDT)
**Key Pattern**: Node-local counters with distributed reads
- Each node only increments its own counter
- Reads gather and sum from all nodes
- Caching for partition tolerance

## Challenge #5: Kafka-Style Log
**Implementation Strategy**:
- **Single-Node**: Map-based offset storage with append/commit operations
- **Multi-Node**: Three bucket types (latest offset, committed offset, messages)
- **Efficient**: Consolidated storage with consistent hashing and reduced CAS operations

## Challenge #6: Transactions
**Consistency Levels**:
- **Single-Node**: Basic in-memory transactions
- **Read Uncommitted**: Prevented dirty writes with local locking and async replication
- **Read Committed**: Added versioned values to prevent dirty reads while maintaining availability

## Core Distributed Systems Patterns Used

1. **Coordination Avoidance**: Generating IDs without consensus (Challenge #2)
2. **Gossip Protocol**: Epidemic-style message propagation (Challenge #3)
3. **Conflict-Free Replicated Data Types**: Per-node counters with merge function (Challenge #4)
4. **Sharding**: Consistent hashing for message routing (Challenge #5)
5. **Optimistic Concurrency Control**: Compare-and-swap with retries (Challenge #5)
6. **Multi-Version Concurrency Control**: Versioned values for isolation (Challenge #6)
7. **Eventual Consistency**: Async replication with eventual convergence (Multiple challenges)

These principles form the backbone of my solutions, balancing performance, consistency, and availability according to each challenge's requirements.
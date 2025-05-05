# Challenge #5b: Multi-Node Kafka-Style Log
> [desc](https://fly.io/dist-sys/5b/)

Took our single-node Kafka and went distributed. Swapped in-memory maps for linearizable KV store and got a rock-solid multi-node log. Turns out atomic CAS + retry loops = distributed coordination gold.

### How it Works

#### Distributed Coordination
- Lin-KV store as source of truth (goodbye maps, hello linearizability)
- Atomic offset reservation via CAS retry loops
- Smart key design: `msg:{topic}:{offset}`, `next:{topic}`, `commit:{topic}`
- Careful type handling with `toInt()` to handle JSON numeric quirks

### Why This Works

1. **Linearizable Offset Allocation**: CAS loop guarantees unique, monotonic offsets
2. **Optimistic Concurrency**: No locks, just retry on conflict = max throughput
3. **Strong Consistency**: Lin-KV gives us the guarantees we need where it matters
4. **Sparse Log Support**: Inherent in our key design, gaps just work

### metrics

* Messages / op: 14.67
* Availability: 0.9994939
* Throughput avg: 790 ops/s
* Server-side messages / op: 12.2

### Trade-offs

**Pros:**
- Zero coordination overhead between nodes (lin-kv does the heavy lifting)
- Natural fault tolerance (nodes are stateless, KV store persists)
- Scales to arbitrary node count

**Cons:**
- Network-bound performance (RPC per operation)
- No batching for high-throughput scenarios
- Linear polling cost for large offset ranges
- No custom replication factor control

Could optimize with batch operations or client-side caching, but this clean approach crushes the spec. Lin-kv proving once again that the right primitive makes distributed systems borderline trivial.

Next try to improve performance.
## Challenge #4: G-Counter (Eventually Consistent Counter)

> [desc](https://fly.io/dist-sys/4/)

Built a distributed counter that survives network partitions using node-local counters and eventual consistency. No fancy CRDTs needed - just good old divide and conquer.

### How it Works

#### Local First, Global Later
- Each node owns its chunk of the counter
- Adds are super fast (just local increment)
- Reads aggregate from all nodes concurrently
- Network partition? No problem, nodes keep counting

#### Implementation Bits
- `localValue` with mutex for thread safety
- KV store for persistence
- Concurrent reads using goroutines
- Simple message types: add, read, local

### Why This Works
1. **Partition Friendly**: Nodes keep working even when isolated
2. **Eventually Consistent**: Global reads converge once network heals
3. **Fast Writes**: No coordination needed for increments
4. **Clean Code**: ~150 lines tells the whole story

## Trade-offs
**Pros:**
- Simple to reason about
- Naturally partition tolerant 
- Local writes are instant
- No complex state merging

**Cons:**
- Reads hit all nodes
- Network issues = incomplete reads
- Not strictly consistent (but that's fine)

Could've gone full CRDT, but letting each node count its own stuff just works. Plus, Go's goroutines make those concurrent reads feel like cheating. 

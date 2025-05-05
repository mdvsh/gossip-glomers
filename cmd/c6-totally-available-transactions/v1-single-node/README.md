# Challenge #6a: Single-Node Totally Available Transactions
> [desc](https://fly.io/dist-sys/6a/)

 a single-node key/value store with transaction support. Turns out you can get pretty far with a map and a mutex.

### How it Works

#### Transaction Processing
- In-memory map for the KV store
- Process read/write ops sequentially within txn
- Grab a lock, do the work, release when done
- Read returns nil for missing keys (spec says that's cool)

### Why This Works
1. **Atomic Transactions**: Lock the whole thing, do all ops, unlock
2. **Simple Semantics**: Read what's there, write what you're told
3. **Total Availability**: Single node = no coordination needed
4. **Stateful**: Everything lives in memory, so it's fast

### Trade-offs
**Pros:**
- Dead simple implementation
- Fast in-memory operations
- No need for complex transaction logic in single-node
- Read-after-write consistency comes for free

**Cons:**
- Coarse-grained locking (could be more granular)
- Unbounded memory growth
- No persistence (crashes = data loss)
- Won't scale to multi-node without more work

Could optimize with more granular locking or a proper storage backend, but honestly this passes the tests. Curious how the multi-node version will complicate things...
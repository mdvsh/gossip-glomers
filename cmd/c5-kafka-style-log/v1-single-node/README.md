# Challenge #5a: Single-Node Kafka-Style Log
> [desc](https://fly.io/dist-sys/5a/)

Built a single-node log service that speaks Kafka. Handles sparse offsets and never loses writes, all while keeping things monotonic. Map-based storage made this surprisingly clean.

### How it Works

#### Message Storage
- Map for offset -> message (sparse offsets for free)
- Thread-safe ops with RWMutex
- Monotonic offset counter separate from storage
- No assumptions about gaps or contiguity

### Why This Works
1. **Monotonic Offsets**: nextOffset counter never goes backwards
2. **No Lost Writes**: Map storage means gaps are fine but nothing disappears
3. **Thread Safety**: RWMutex keeps concurrent ops clean
4. **Simple Polling**: Sequential scan from requested offset just works

### Trade-offs
**Pros:**
- Natural sparse offset handling 
- Clean concurrent access patterns
- Simple offset management
- Easy to reason about safety properties

**Cons:**
- Unbounded memory usage (but spec doesn't care)
- Sequential scans for polls
- No cleanup mechanism
- Full copy on response building

Can go fancy with skip lists or cleanup strategies, but keeping it simple with maps just worked.
Next, multi-node!
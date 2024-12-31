## Challenge #3b: Multi-Node Broadcast

> [desc](https://fly.io/dist-sys/3b/)

Built on our single-node code, now we're actually spreading messages across multiple nodes! Kept it super simple but effective - just tell everyone about new messages.

### What it Does
Instead of getting fancy with topology or complicated gossip protocols, went with the most straightforward solution:
- When a node gets a new message, tell literally everyone else
- Skip the node that sent us the message (no echo chambers!)
- Skip ourselves (we already know the message)
- Fire-and-forget sends for speed

### How it Works

#### Message Propagation

Each new message gets blasted out to all other nodes (except the sender). Using goroutines to send messages concurrently keeps things snappy.

#### Message Storage

- Using a map for O(1) duplicate detection
- Empty struct value saves memory (0 bytes!)
- Clean lock handling to prevent races

#### Why This Works
1. **Full Connectivity**: Every node knows about every other node
2. **Direct Communication**: No need to route through intermediaries
3. **Fast Propagation**: New messages spread in one hop
4. **Duplicate Prevention**: Map-based deduplication keeps us efficient

## Trade-offs
**Pros:**
- Dead simple implementation
- fast message propagation (one hop)
- No topology management needed
- Easy to reason about

**Cons:**
- Not network efficient (O(n) messages per broadcast)
- More network traffic than necessary
- Might not scale well to huge clusters

But for our needs here ig, simple is better than complex. This does the job reliably and the code is clear. We'll optimize if we need to in 3d and 3e!
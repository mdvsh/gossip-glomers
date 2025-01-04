## Challenge #3d: Efficient Broadcast 

> [desc](https://fly.io/dist-sys/3d/)

Built on our fault-tolerant broadcast, now we're handling bigger clusters and network latency! wild part? the simple periodic sync solution from 3c turned out to be efficient.

### What it Does
Instead of trying to optimize topology or create complex protocols, kept our dead simple approach:
- Every 500ms, tell everyone what you know
- Let natural batching do the heavy lifting
- Still handles network partitions (just takes a bit longer to sync)
- Surprisingly efficient at node scale

### How it Works

#### Message Batching
The magic happens automatically:
- Test sends 100 messages/second
- We sync every 500ms
- ~50 messages batch together naturally
- Cost gets amortized across all messages

#### Performance 
With 25 nodes and 100ms latency:
- ~23 messages per op (way under limit of 30)
- ~300ms median latency (limit was 400ms)
- ~500ms p99 latency (still under the 600ms limit)

#### Why This Works
1. **Natural Batching**: Multiple messages share the sync cost
2. **Good Timing**: 500ms interval hits the sweet spot
3. **Still Fault-Tolerant**: Keeps working through partitions
4. **Simple Scaling**: Works even better than our naive broadcast!

## Trade-offs
**Pros:**
- _still_ Simple
- Naturally efficient batching
- Still handles partitions
- Actually scales pretty well

**Cons:**
- Luck played a part in the timing
- Might need tuning for different workloads
- Not the most sophisticated approach

But sometimes one gets lucky and the simple solution turns out to be secretly smart.
I'll take it for now good knowing that no solution is perfect, and it boils down to trade-offs and balances.
It was interesting to play with latency and the nodes with this. Reducing ticker interval was at the cost of increased messages-per-op but lower median, p99, and max latencies (~300 and ~500ms resp. even).
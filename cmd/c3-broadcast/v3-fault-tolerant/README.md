## Challenge #3c: Fault-Tolerant Broadcast

> [desc](https://fly.io/dist-sys/3c/)

Built on our multi-node broadcast, but now we're handling network partitions. Instead of getting fancy with message queues and retry logic, I tried a simple solution and it just works.

### What it Does
just periodically tell everyone everything you know.
- Each node syncs its entire message set every 500ms
- No tracking failed messages
- No complex retry logic
- Just keep sharing until everyone has everything

### How it Works

#### Periodic State Sync
Every 500ms, each node:
1. Grabs its current message set
2. Sends it to every other node

and 

#### Message Storage
as before, we're using a map for O(1) duplicate detection.

### Why This Works 
1. **Simple Gossip Protocol**
2. **Self-Healing**: Nodes automatically catch up after partitions heal
3. **Eventually Consistent**: Messages reach everyone... eventually
4. **Zero Tracking**: No need to remember what failed or succeeded

## Trade-offs
**Pros:**
- Simple
- Self-healing by design
- No complex state tracking
- Easy to reason about

**Cons:**
- More network traffic than necessary
- Repeated sending of old messages
- Not super efficient in space/time, as we're sending everything we know over network

Next up: 3d/e will make us actually think about efficiency... but for now, this is fine. 
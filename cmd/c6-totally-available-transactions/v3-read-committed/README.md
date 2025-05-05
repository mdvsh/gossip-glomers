# Challenge #6c: Totally-Available, Read Committed Transactions
> [desc](https://fly.io/dist-sys/6c/)

Leveling up our transaction system with an extra guarantee: no dirty reads.
We've upgraded our consistency model while keeping total availability.

### How it Works

#### Versioned Key-Value Store
- Each key stores a list of versioned values
- Values track their committed state
- Reads only see committed values
- Writes live in transaction until commit time

#### Isolation & Replication
- Reads only see the latest committed value
- Writes are staged in the transaction object
- Atomic commit applies all transaction writes
- Committed writes are replicated async to other nodes
- Support for transaction abort (required by spec)

### Why This Works
1. **Versioned Values**: Only committed values are visible to reads
2. **Transaction Staging**: Buffer all writes until commit
3. **Atomic Commits**: All writes become visible at once
4. **Async Replication**: Fire-and-forget keeps us totally available
5. **No Wait-for-Ack**: We keep going even if nodes are partitioned

### Trade-offs
**Pros:**
- Still relatively simple design
- Remains totally available during partitions
- Prevents all anomalies (G0, G1a, G1b, G1c) as required
- No distributed coordination needed
- Final form! Complete consistency path from single node → read-uncommitted → read-committed

**Cons:**
- Storage overhead from versioning
- Still weak consistency (no repeatable reads)
- Values eventually replicate, but no guarantees when
- Partition recovery requires no special handling
- Transaction ID generation is ultra-simple (but works!)

The subtle art of keeping enough consistency to pass the test while staying totally available.
i think this was the last challenge, so ill see you when i see you!

---

ref:
Read Committed prohibits:
- G0 (dirty write): transaction write conflicts
- G1a (aborted read): reading data from aborted transactions
- G1b (intermediate read): reading partial/intermediate transaction results
- G1c (circular information flow): transactions creating circular dependencies
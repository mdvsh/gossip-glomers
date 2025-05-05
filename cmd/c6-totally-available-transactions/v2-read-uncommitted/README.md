# Challenge #6b: Totally-Available, Read Uncommitted Transactions 
> [desc](https://fly.io/dist-sys/6b/)

Multi-node transactions with the weakest consistency model known to humanity. 
Can replication really be this easy?

### How it Works

#### Transaction Processing + Replication
- In-memory map stays from v1, still no persistence
- Process writes locally first, broadcast async to other nodes
- No need for consensus, just fire-and-forget writes
- Reads still return nil for missing keys (feature, not bug!)
- Only blocks G0 (dirty writes) via local locking

### Why This Works
1. **Async Replication**: Fire off writes, don't wait for acknowledgments
2. **Total Availability**: Network partitions? No problem, keep operating
3. **Global Dirty Write Prevention**: Local locks prevent G0 within nodes
4. **Topology Tracking**: Learn about all nodes for broadcasting
5. **No Cost Reads**: Read whatever's locally available, it's valid

### Trade-offs
**Pros:**
- Still pretty simple implementation
- No coordination overhead for writes
- No need to wait for replicas
- Totally available even during partitions
- Will pass the tests! 

**Cons:**
- Weak consistency guarantees - basically anything goes
- Only protects against G0, allowing all kinds of fun anomalies
- No guarantees about writes making it to other nodes
- Nodes can diverge during partitions
- Real read uncommitted would need write locking between nodes (overkill for spec)

Read uncommitted is amusing - it's the "consistency for people who hate consistency" model. 
Looking forward to the read committed challenge where we'll have to prevent dirty reads too.

---

ref:

> G0 (dirty write): a cycle of transactions linked by write-write dependencies. For instance, transaction T1 appends 1 to key x, transaction T2 appends 2 to x, and T1 appends 3 to x again, producing the value [1, 2, 3].
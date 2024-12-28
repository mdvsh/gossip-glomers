## Challenge #2: Unique ID Generation

> [desc](https://fly.io/dist-sys/2/)

Built a distributed unique ID generator that keeps working even when the network gets funky (partition tolerant)

### How it Works
Each ID is a combo of:
- Timestamp (nanoseconds) - gives us time ordering
- Node ID - makes sure different nodes don't clash
- Random bytes - prevents collisions if we generate IDs at the exact same nanosecond

Why this approach is cool:
- No coordination needed between nodes (works in partitions)
- IDs are roughly time-ordered (good for debugging)
- Basically zero chance of collisions
- Fast! Just local operations

### Testing It
```bash
go build
./maelstrom test -w unique-ids --bin ./unique --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
```

### Why It's Partition Tolerant
Each node works independently - no need to talk to other nodes to generate IDs. Even if the network splits, each part keeps generating valid, unique IDs. When the partition heals, all IDs are still unique because of the node ID component.

### Trade-offs
- IDs are strings (could be more compact)
- Not perfectly time-ordered across nodes (would need coordination for that)
- ID length varies with node ID length

But for our use case, these are fine - we prioritized availability and simplicity.
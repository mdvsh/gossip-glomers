## Challenge #3a: Single-Node Broadcast

> [desc](https://fly.io/dist-sys/3a/)

Building a broadcast system that'll eventually gossip messages between nodes. Starting simple with just one node to get the basic message handling solid.

### What it Does
- Stores broadcast messages in memory
- Handles three main operations:
  - `broadcast`: Takes a message (integer) and saves it
  - `read`: Returns all messages we've seen
  - `topology`: Gets network topology info (not used yet but will be key for gossip)

### How it Works
Using a server struct with:
- Thread-safe message storage (slice + mutex)
- Clean message type definitions
- Separate handlers for each operation

Key design choices:
- Used slice instead of map since we just need to store unique integers
- RWMutex for better read performance (multiple readers can access simultaneously)
- Message copying in read handler to prevent external modifications

---

### Trade-offs
- Using slice for storage (simple but O(n) lookup)
- Copying full message list on reads (safe but not memory efficient)
- Storing topology even though we don't use it yet (preparing for 3b)

These are fine for single-node but we'll need to revisit some choices when we add multi-node support.
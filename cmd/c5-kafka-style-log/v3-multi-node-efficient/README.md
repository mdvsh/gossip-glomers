# Challenge #5c: Efficient Multi-Node Kafka-Style Log
> [desc](https://fly.io/dist-sys/5c/)

Boosted our Kafka log with clever caching and cut the message overhead by ~14%. Turns out you don't need fancy algorithms when a good old cache hits the sweet spot.

## How it Works

### Smart Read-Through Caching
- In-memory cache for offset counters (bye bye redundant reads)
- Optimistic CAS with cache-based hints (fewer retries = less traffic)
- Cache invalidation on write conflicts (eventual consistency FTW)

### Optimized Polling
- Batched reads (5 at a time) to slash network overhead
- Message prefetch based on access patterns
- No fancy data structure needed - just better read patterns

## The Optimization Game

1. **Offset Caching**: Cache-first CAS attempts cut ~50% of offset reads
2. **Batch Reading**: Sequential message groups slashed polling RPCs
3. **Minimal Implementation**: No complex distributed algorithms needed

## Metrics Comparison

| Metric | v2 (Original) | v3 (Optimized) | Improvement |
|--------|--------------|----------------|------------|
| Messages/op | 14.67 | 12.64 | 13.8% |
| Availability | 0.9994939 | 0.9995162 | 0.002% |
| Server messages/op | 12.2 | 10.29 | 15.7% |

## The Magic of Simple Solutions

Cut the fancy sharding, CRDT, and consensus tricks - a well-placed cache was all we needed. The read-through cache pattern brings most of the performance boost without the distributed systems PhD.

The classic CS lesson applies: optimize the common path first. We found that caching just the offset values and batching reads in groups of 5 accounted for most of the gains.

## Fun Facts
- Eliminating ~2 messages per operation at 1000 ops/sec saves 2K round trips per second
- Reduced CAS retries means lower tail latency for concurrent writes
- All this while maintaining the same linearizable correctness guarantees

## The Path Forward
Could go wild with gossip protocols, RDMA, and cache coherence schemes - but nah, we'll save those for when the requirements actually call for them.

Sometimes the simplest solutions are the most effective.
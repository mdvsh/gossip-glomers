## Challenge #3e: Even More Efficient Broadcast!

> [desc](https://fly.io/dist-sys/3e/)

Built on our already efficient broadcast, just cranked up the efficiency by, waiting longer? Yeah, turns out sometimes being lazy is the smart move.

### What Changed
literally just bumped the sync interval to 1s from 500ms. that's it. no fancy algorithms, no complex napkin math, just:
```go
ticker := time.NewTicker(time.Second)
```

### How it Works (Even Better)

#### Even More Batching
Same principle as before but better:
- Test still sends 100 messages/second
- Now we sync every 1s instead of 500ms
- Cost gets even more amortized across messages
- Ended up with only ~13 messages per op (crushed the test limit of 20)

#### Perf Numbers
Still with 25 nodes and 100ms latency:
- ~13 messages per op (way down from our previous ~23)
- ~564ms median latency (limit was 1s, so ton of headroom)
- ~1077ms max latency (not even close to the 2s limit)

### Why This Works Better
1. **More Patient Batching**: Waiting longer = bigger batches = fewer messages
2. **Still Good Timing**: 1s interval hits an even sweeter spot
3. **Same Simple Code**: Didn't have to add any complexity
4. **Free Performance**: Got better numbers by doing less work

## Trade-offs
**Pros:**
- _still_ Simple (sensing a theme here?)
- Even more efficient batching
- Same fault tolerance
- Actually got better by being lazier

**Cons:**
- Higher latency (but still well within limits)
- More "wait and see" in the system
- Still not the most sophisticated approach

Sometimes the best optimization is just tweaking a number.
there's probably fancier ways to do this with topology optimization or smart gossip protocols, but if it works, it works... And this definitely works. also, happy new year from time of writing.
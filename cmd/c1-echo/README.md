## Challenge #1: Echo

> [desc](https://fly.io/dist-sys/1/)

First challenge in the [series](https://fly.io/dist-sys/) - pretty straightforward but gets us familiar with how [Maelstrom](https://github.com/jepsen-io/maelstrom) works. Basically just built a service that bounces back messages, but it's a good foundation for the more interesting distributed stuff coming up.

### What it Does
- Takes in messages through STDIN
- Sends them back with an `echo_ok` type
- Keeps track of message IDs so replies match up with requests
- Uses Maelstrom's Go library to handle the boring parts

Even though it's just echo-ing messages back, we're already dealing with:
- JSON message passing between nodes
- Proper request/response cycles
- Message correlation (keeping track of which response goes with which request)

These patterns show up everywhere in distributed systems, so it's good to get them down.

### Testing It
```bash
go build # in cmd/<proj> dir
# run binary
./maelstrom test -w echo --bin ./echo --node-count 1 --time-limit 10
```

Everything passed:
- Handled 47 messages without any fails
- Perfect availability (didn't drop anything)
- Clean 2 messages per op (just request -> response)

Kept the full test output in `testdata/` because why not.
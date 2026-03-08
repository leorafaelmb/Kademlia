# Kademlia

A Kademlia DHT implementation in Go, following [BEP 5](https://www.bittorrent.org/beps/bep_0005.html) for BitTorrent peer discovery. Built as a standalone library for the [JellyTorrent](https://github.com/leorafaelmb/JellyTorrent) BitTorrent client.

Zero external dependencies.

## Usage

```go
import Kademlia "github.com/leorafaelmb/Kademlia"

// Create and start a DHT node
dht, err := Kademlia.New(
    Kademlia.WithPort(6881),
    Kademlia.WithLogger(slog.Default()),
)
if err != nil {
    log.Fatal(err)
}
defer dht.Close()

// Populate the routing table
ctx := context.Background()
dht.Bootstrap(ctx)

// Find peers for a torrent
infoHash := [20]byte{ /* SHA1 of torrent info dict */ }
peers, err := dht.GetPeers(ctx, infoHash)

// Announce that you are downloading/seeding a torrent
dht.Announce(ctx, infoHash, 6881)
```

## Architecture

```
Kademlia/
  dht.go              Public API: New, Bootstrap, GetPeers, Announce, Close
  config.go           Config struct + functional options
  peerstore.go        In-memory peer storage with TTL expiration
  internal/
    nodeid/            20-byte node ID, XOR distance, prefix length
    routing/           K-bucket routing table (K=8, 160 buckets, LRU eviction)
    krpc/              KRPC protocol over UDP (message encoding, transactions, server)
    token/             HMAC-SHA1 token generation/validation with rotating secrets
    lookup/            (reserved for future refactoring)
```

### Key components

- **NodeID** - Random 160-bit identifier. XOR distance determines routing and storage responsibility.
- **Routing Table** - 160 k-buckets indexed by shared prefix length. Nodes are ordered by last-seen time (LRU). `FindClosest` returns the K nearest nodes to any target by XOR distance.
- **KRPC Server** - UDP read loop that dispatches incoming queries to handlers and matches responses to pending transactions.
- **Token Manager** - HMAC-SHA1 tokens tied to querier IP, with two rotating secrets (current + previous) for graceful expiration.
- **Peer Store** - Maps info hashes to peer addresses with configurable TTL (default 30 min).

### Protocol

Four RPCs over UDP using bencoded messages:

| Method | Purpose |
|---|---|
| `ping` | Liveness check, returns node ID |
| `find_node` | Returns K closest nodes to a target ID |
| `get_peers` | Returns peers for an info hash, or K closest nodes + token |
| `announce_peer` | Registers caller as a peer (requires valid token) |

### Maintenance

Background goroutines handle:

- **Token rotation** every 5 minutes
- **Peer expiration** every 5 minutes (removes entries older than TTL)
- **Routing table refresh** every 15 minutes (find_node on random targets)

## Configuration

| Option | Default | Description |
|---|---|---|
| `WithPort(port)` | 6881 | UDP listen port |
| `WithBootstrapNodes(nodes)` | 3 public routers | Initial nodes to contact |
| `WithAlpha(n)` | 3 | Concurrent queries per lookup round |
| `WithLogger(l)` | `slog.Default()` | Structured logger |

## Build

```bash
go build ./...
go test ./...
```

## References

- [BEP 5: DHT Protocol](https://www.bittorrent.org/beps/bep_0005.html)
- [Kademlia: A Peer-to-peer Information System Based on the XOR Metric](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf)

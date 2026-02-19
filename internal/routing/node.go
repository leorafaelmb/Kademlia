package routing

import (
	"github.com/leorafaelmb/Kademlia/internal/nodeid"
	"net/netip"
	"time"
)

type Node struct {
	ID       nodeid.NodeID
	Addr     netip.AddrPort
	LastSeen time.Time
}

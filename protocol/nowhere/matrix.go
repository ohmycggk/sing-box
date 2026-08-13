package nowhere

import (
	"fmt"

	"github.com/ohmycggk/nowhere-go/carrier/tcptls"
	"github.com/ohmycggk/nowhere-go/wire"
)

const (
	DefaultUp       = "udp"
	DefaultDown     = "udp"
	DefaultPoolSize = 5
	MaxPoolSize     = tcptls.MaxPoolSize
)

// Matrix is the resolved outbound carrier selection (up/down) plus pool policy.
type Matrix struct {
	Up          string
	Down        string
	Pool        int
	Asymmetric  bool
	NeedsQUIC   bool
	NeedsTCP    bool
	poolWarning string
}

// ResolveMatrix applies Nowhere outbound defaults and pool rules.
//
// Rules (protocol / merge plan §4.2):
//   - up/down both empty → udp/udp
//   - exactly one set → error
//   - pool only effective for tcp/tcp (default 5); any UDP carrier forces pool=0
func ResolveMatrix(up, down string, pool *int) (Matrix, error) {
	if up == "" && down == "" {
		up, down = DefaultUp, DefaultDown
	}
	if (up == "") != (down == "") {
		return Matrix{}, fmt.Errorf("nowhere: up and down must both be set or both omitted")
	}
	if !isCarrier(up) || !isCarrier(down) {
		return Matrix{}, fmt.Errorf("nowhere: invalid carrier selector up=%q down=%q", up, down)
	}

	m := Matrix{
		Up:         up,
		Down:       down,
		Asymmetric: up != down,
		NeedsQUIC:  up == "udp" || down == "udp",
		NeedsTCP:   up == "tcp" || down == "tcp",
	}
	if up == "tcp" && down == "tcp" {
		if pool == nil {
			m.Pool = DefaultPoolSize
		} else if *pool < 0 {
			return Matrix{}, fmt.Errorf("nowhere: pool must be >= 0")
		} else if *pool > MaxPoolSize {
			m.Pool = MaxPoolSize
			m.poolWarning = fmt.Sprintf("nowhere: pool %d exceeds maximum %d; using %d", *pool, MaxPoolSize, MaxPoolSize)
		} else {
			m.Pool = *pool
		}
	} else {
		// Rust parses pool only for tcp/tcp. Every matrix involving UDP ignores
		// it, including values that would be invalid for a TCP pool.
		if pool != nil && *pool != 0 {
			m.poolWarning = fmt.Sprintf("nowhere: pool is only effective for tcp/tcp; ignoring configured value %d", *pool)
		}
		m.Pool = 0
	}
	return m, nil
}

func isCarrier(s string) bool { return s == "tcp" || s == "udp" }

// RequiresFlowEnvelope reports whether asymmetric FLOW_OPEN/FLOW_ATTACH must be used.
func (m Matrix) RequiresFlowEnvelope() bool { return m.Asymmetric }

// FlowKindForNetwork maps a user flow to wire.FlowKind.
func FlowKindForNetwork(network string) (wire.FlowKind, error) {
	switch network {
	case "tcp":
		return wire.FlowKindTCP, nil
	case "udp":
		return wire.FlowKindUDP, nil
	default:
		return 0, fmt.Errorf("nowhere: invalid flow kind %q", network)
	}
}

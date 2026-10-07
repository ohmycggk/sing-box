package option

import (
	"github.com/sagernet/sing/common/json/badoption"
)

// NowhereOutboundOptions is the SingBox Nowhere client configuration.
type NowhereOutboundOptions struct {
	DialerOptions
	ServerOptions
	Password string `json:"password,omitempty"`
	// Pin is the leaf certificate SHA-256 (lowercase hex). When set it overrides
	// SNI/chain verification for Nowhere TLS (TCP and QUIC).
	Pin  string `json:"pin,omitempty"`
	Up   string `json:"up,omitempty"`
	Down string `json:"down,omitempty"`
	Mux  *int   `json:"mux,omitempty"`
	Pool *int   `json:"pool,omitempty"`
	// Morph enables the Nowhere 2 Morph keyed transform for every carrier of
	// this endpoint (64-byte TCP prelude below TLS, directional ChaCha20 keys
	// below QUIC). There is no negotiation: both ends of a hop must agree.
	Morph                 bool               `json:"morph,omitempty"`
	PrewarmOnStart        bool               `json:"prewarm_on_start,omitempty"`
	MaxConcurrentDials    *int               `json:"max_concurrent_dials,omitempty"`
	WarmBackoffInitial    badoption.Duration `json:"warm_backoff_initial,omitempty"`
	WarmBackoffMax        badoption.Duration `json:"warm_backoff_max,omitempty"`
	QUICCongestionControl string             `json:"quic_congestion_control,omitempty"`
	OutboundTLSOptionsContainer
	QUICOptions
}

// NowhereNextOptions configures native Portal chaining on a Nowhere inbound:
// authenticated flows are forwarded to another Nowhere Portal instead of the
// local router. ServerName/Pin mirror the outbound's tls.server_name and pin.
type NowhereNextOptions struct {
	ServerOptions
	Password   string `json:"password,omitempty"`
	Up         string `json:"up,omitempty"`
	Down       string `json:"down,omitempty"`
	Mux        *int   `json:"mux,omitempty"`
	Pool       *int   `json:"pool,omitempty"`
	ServerName string `json:"server_name,omitempty"`
	// Pin is the leaf certificate SHA-256 (lowercase hex). When set it overrides
	// SNI/chain verification for the chained Portal TLS (TCP and QUIC).
	Pin string `json:"pin,omitempty"`
	// Morph overrides Morph toward the next Portal. Nil inherits the inbound's
	// morph setting.
	Morph *bool `json:"morph,omitempty"`
	// Dial4 binds the IPv4 source address used when dialing the next Portal
	// over IPv4 ("auto" or an empty value leaves the source address unset).
	// Dial6 does the same for IPv6. Both mirror the Rust portal URL parameters
	// dial4/dial6 and are mutually exclusive with dial, which has no equivalent
	// here: sing-box already exposes inet4_bind_address / inet6_bind_address on
	// the outbound's shared dial fields.
	Dial4 string `json:"dial4,omitempty"`
	Dial6 string `json:"dial6,omitempty"`
}

// NowhereInboundOptions is the SingBox Nowhere Portal configuration.
// TLS is always the native SingBox TLS object; plaintext and historical
// numeric tls/crt/key compatibility forms are not accepted.
type NowhereInboundOptions struct {
	ListenOptions
	Password string `json:"password,omitempty"`
	// Morph enables the Nowhere 2 Morph keyed transform for every carrier of
	// this endpoint. There is no negotiation: clients must use the same
	// setting.
	Morph                         bool                `json:"morph,omitempty"`
	Network                       NetworkList         `json:"network,omitempty"`
	QUICCongestionControl         string              `json:"quic_congestion_control,omitempty"`
	MaxUnauthenticatedConnections *int                `json:"max_unauthenticated_connections,omitempty"`
	MaxUnauthenticatedPerSource   *int                `json:"max_unauthenticated_per_source,omitempty"`
	Next                          *NowhereNextOptions `json:"next,omitempty"`
	InboundTLSOptionsContainer
	QUICOptions
}

package option

import (
	"github.com/sagernet/sing/common/byteformats"
	"github.com/sagernet/sing/common/json/badoption"
)

// QUICOptions holds shared QUIC transport tuning for Nowhere inbound/outbound.
type QUICOptions struct {
	IdleTimeout             badoption.Duration      `json:"idle_timeout,omitempty"`
	KeepAlivePeriod         badoption.Duration      `json:"keep_alive_period,omitempty"`
	StreamReceiveWindow     byteformats.MemoryBytes `json:"stream_receive_window,omitempty"`
	ConnectionReceiveWindow byteformats.MemoryBytes `json:"connection_receive_window,omitempty"`
	MaxConcurrentStreams    int                     `json:"max_concurrent_streams,omitempty"`
	InitialPacketSize       int                     `json:"initial_packet_size,omitempty"`
	DisablePathMTUDiscovery bool                    `json:"disable_path_mtu_discovery,omitempty"`
}

// NowhereOutboundOptions is the SingBox Nowhere client configuration.
type NowhereOutboundOptions struct {
	DialerOptions
	ServerOptions
	Password string `json:"password,omitempty"`
	// Pin is the leaf certificate SHA-256 (lowercase hex). When set it overrides
	// SNI/chain verification for Nowhere TLS (TCP and QUIC).
	Pin                   string             `json:"pin,omitempty"`
	Up                    string             `json:"up,omitempty"`
	Down                  string             `json:"down,omitempty"`
	Pool                  *int               `json:"pool,omitempty"`
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
	Pool       *int   `json:"pool,omitempty"`
	ServerName string `json:"server_name,omitempty"`
	// Pin is the leaf certificate SHA-256 (lowercase hex). When set it overrides
	// SNI/chain verification for the chained Portal TLS (TCP and QUIC).
	Pin string `json:"pin,omitempty"`
}

// NowhereInboundOptions is the SingBox Nowhere Portal configuration.
// TLS is always the native SingBox TLS object; plaintext and historical
// numeric tls/crt/key compatibility forms are not accepted.
type NowhereInboundOptions struct {
	ListenOptions
	Password                      string              `json:"password,omitempty"`
	Network                       NetworkList         `json:"network,omitempty"`
	QUICCongestionControl         string              `json:"quic_congestion_control,omitempty"`
	MaxUnauthenticatedConnections *int                `json:"max_unauthenticated_connections,omitempty"`
	MaxUnauthenticatedPerSource   *int                `json:"max_unauthenticated_per_source,omitempty"`
	Next                          *NowhereNextOptions `json:"next,omitempty"`
	InboundTLSOptionsContainer
	QUICOptions
}

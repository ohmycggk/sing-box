package server

import (
	gonowhere "github.com/ohmycggk/nowhere-go/server"
	"github.com/ohmycggk/nowhere-go/wire"
	"github.com/sagernet/sing-box/option"
	N "github.com/sagernet/sing/common/network"
)

// Type aliases for thin re-export of nowhere-go/server.
type (
	Config       = gonowhere.Config
	Handler      = gonowhere.Handler
	Upstream     = gonowhere.Upstream
	CloseHandler = gonowhere.CloseHandler
	QuicConn     = gonowhere.QuicConn
	QuicListener = gonowhere.QuicListener
	QuicStream   = gonowhere.QuicStream
)

// IsReported reports whether the protocol core already emitted the error.
var IsReported = gonowhere.IsReported

// NewConfig builds nowhere-go Config from SingBox inbound options.
func NewConfig(options option.NowhereInboundOptions) (*Config, error) {
	credentials, err := wire.NewCredentials(options.Password)
	if err != nil {
		return nil, err
	}
	rawNetworks := options.Network.Build()
	networks := make([]gonowhere.Network, 0, len(rawNetworks))
	for _, network := range rawNetworks {
		switch network {
		case N.NetworkTCP:
			networks = append(networks, gonowhere.NetworkTCP)
		case N.NetworkUDP:
			networks = append(networks, gonowhere.NetworkUDP)
		default:
			networks = append(networks, gonowhere.Network(network))
		}
	}
	limits := gonowhere.Limits{}
	if options.MaxUnauthenticatedConnections != nil {
		limits.MaxUnauthenticatedConnections = *options.MaxUnauthenticatedConnections
	}
	if options.MaxUnauthenticatedPerSource != nil {
		limits.MaxUnauthenticatedPerSource = *options.MaxUnauthenticatedPerSource
	}
	alpn := wire.DefaultALPN
	if options.TLS != nil && len(options.TLS.ALPN) == 1 {
		alpn = string(options.TLS.ALPN[0])
	}
	return gonowhere.NewConfig(gonowhere.ConfigOptions{
		Credentials: credentials,
		ALPN:        alpn,
		Networks:    networks,
		Limits:      limits,
	})
}

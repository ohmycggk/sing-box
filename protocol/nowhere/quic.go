//go:build with_quic

package nowhere

import (
	"github.com/ohmycggk/nowhere-go/carrier"
	quicpkg "github.com/sagernet/sing-box/protocol/nowhere/carrier/quic"
)

const quicIncluded = true

// QUICConfig remains available to with_quic host integrations while the
// untagged facade stays free of concrete QUIC dependencies.
type QUICConfig = quicpkg.QUICConfig

func newQuicBackend(options quicBackendOptions) carrier.QuicBackend {
	return NewQuicBackend(&quicpkg.QUICConfig{
		Context:           options.context,
		Addr:              options.address,
		ServerName:        options.serverName,
		TLSConfig:         options.tlsConfig,
		QUICConfig:        quicpkg.BuildQUICConfig(options.quicOptions),
		Dialer:            options.dialer,
		CongestionControl: options.congestionControl,
		Observer:          options.observer,
	})
}

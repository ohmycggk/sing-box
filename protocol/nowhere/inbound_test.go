package nowhere

import (
	"context"
	"testing"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/stretchr/testify/require"
)

func TestInboundImplementsInterfaceUpdateListener(t *testing.T) {
	t.Parallel()
	var _ adapter.InterfaceUpdateListener = (*Inbound)(nil)
}

func TestNewInboundRejectsUnknownQUICCongestionControlForTCPOnly(t *testing.T) {
	t.Parallel()
	logger := log.NewNOPFactory().Logger()
	_, err := NewInbound(context.Background(), nil, logger, "nw", option.NowhereInboundOptions{
		Password:              "secret",
		QUICCongestionControl: "BBR",
		InboundTLSOptionsContainer: option.InboundTLSOptionsContainer{
			TLS: &option.InboundTLSOptions{Enabled: true},
		},
	})
	require.ErrorContains(t, err, "unknown quic congestion control")
}

func TestInboundInterfaceUpdatedTCPOnlyIsNoop(t *testing.T) {
	t.Parallel()
	in := &Inbound{
		logger:    log.NewNOPFactory().Logger(),
		enableUDP: false,
	}
	require.NotPanics(t, func() { in.InterfaceUpdated(context.Background()) })
}

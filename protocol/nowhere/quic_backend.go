//go:build with_quic

package nowhere

import (
	"context"
	"net"

	nquic "github.com/ohmycggk/nowhere-go/carrier/quic"

	quicpkg "github.com/sagernet/sing-box/protocol/nowhere/carrier/quic"
)

// QuicBackend wires the sing-box QUIC transport primitives into nowhere-go.
type QuicBackend struct {
	client *quicpkg.Client
}

func NewQuicBackend(cfg *quicpkg.QUICConfig) *QuicBackend {
	if cfg == nil {
		return &QuicBackend{}
	}
	configCopy := *cfg
	return &QuicBackend{client: quicpkg.NewClient(&configCopy)}
}

func (b *QuicBackend) AcquireSession(ctx context.Context) (nquic.Session, error) {
	if b == nil || b.client == nil {
		return nil, net.ErrClosed
	}
	return b.client.AcquireSession(ctx)
}

func (b *QuicBackend) InvalidateSession(session nquic.Session) {
	if b != nil && b.client != nil {
		b.client.InvalidateSession(session)
	}
}

func (b *QuicBackend) Close() error {
	if b == nil || b.client == nil {
		return nil
	}
	return b.client.Close()
}

var _ nquic.Backend = (*QuicBackend)(nil)

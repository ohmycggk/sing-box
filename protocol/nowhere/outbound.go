package nowhere

import (
	"context"
	"net"
	"sync"

	corebundle "github.com/ohmycggk/nowhere-go/bundle"
	"github.com/ohmycggk/nowhere-go/carrier"
	"github.com/ohmycggk/nowhere-go/carrier/tcptls"
	"github.com/ohmycggk/nowhere-go/diagnostic"
	"github.com/ohmycggk/nowhere-go/wire"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/nowhere/internal/quicsettings"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

func RegisterOutbound(registry *outbound.Registry) {
	outbound.Register[option.NowhereOutboundOptions](registry, C.TypeNowhere, NewOutbound)
}

var _ adapter.InterfaceUpdateListener = (*Outbound)(nil)

type outboundCarrierOverrides struct {
	tcpDialer      tcptls.TCPDialer
	tlsDialer      tcptls.TLSDialer
	newQUICBackend func() carrier.QuicBackend
}

type outboundCarrierOverridesKey struct{}

func withOutboundCarrierOverrides(ctx context.Context, overrides outboundCarrierOverrides) context.Context {
	return context.WithValue(ctx, outboundCarrierOverridesKey{}, overrides)
}

func outboundOverridesFromContext(ctx context.Context) (outboundCarrierOverrides, bool) {
	overrides, loaded := ctx.Value(outboundCarrierOverridesKey{}).(outboundCarrierOverrides)
	return overrides, loaded
}

type Outbound struct {
	outbound.Adapter
	logger    log.ContextLogger
	observer  diagnostic.Observer
	server    M.Socksaddr
	matrix    Matrix
	mu        sync.Mutex
	bundle    *corebundle.CarrierBundle
	newBundle func() (*corebundle.CarrierBundle, error)
	closed    bool
}

type quicBackendOptions struct {
	context           context.Context
	address           string
	serverName        string
	tlsConfig         tls.Config
	quicOptions       option.QUICOptions
	dialer            N.Dialer
	congestionControl quicsettings.CongestionControl
	observer          diagnostic.Observer
}

func NewOutbound(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.NowhereOutboundOptions) (adapter.Outbound, error) {
	if options.Password == "" {
		return nil, E.New("nowhere: missing password")
	}
	if options.TLS == nil || !options.TLS.Enabled {
		return nil, C.ErrTLSRequired
	}
	tlsOptions, err := normalizeNowhereOutboundTLS(options.TLS)
	if err != nil {
		return nil, err
	}
	options.TLS = tlsOptions
	alpn := string(options.TLS.ALPN[0])
	congestionControl, err := quicsettings.ParseCongestionControl(options.QUICCongestionControl)
	if err != nil {
		return nil, err
	}

	matrix, err := ResolveMatrix(options.Up, options.Down, options.Pool)
	if err != nil {
		return nil, err
	}
	if matrix.poolWarning != "" {
		logger.Warn(matrix.poolWarning)
	}
	if matrix.NeedsQUIC && !quicIncluded {
		return nil, C.ErrQUICNotIncluded
	}

	credentials, err := wire.NewCredentials(options.Password)
	if err != nil {
		return nil, err
	}
	observer := SingObserver{L: logger}

	tlsConfig, err := tls.NewClient(ctx, logger, options.Server, common.PtrValueOrDefault(options.TLS))
	if err != nil {
		return nil, err
	}
	if err := applyNowhereCertificatePin(tlsConfig, options.Pin); err != nil {
		return nil, err
	}
	outboundDialer, err := dialer.NewWithOptions(dialer.Options{
		Context:        ctx,
		Options:        options.DialerOptions,
		RemoteIsDomain: options.ServerIsDomain(),
	})
	if err != nil {
		return nil, err
	}

	server := options.ServerOptions.Build()
	addr := server.String()
	tcpDialer := tcptls.TCPDialer(&socksaddrDialer{dialer: outboundDialer})
	tlsDialer := tcptls.TLSDialer(&singTLSDialer{config: tlsConfig})
	overrides, hasOverrides := outboundOverridesFromContext(ctx)
	if hasOverrides {
		if overrides.tcpDialer != nil {
			tcpDialer = overrides.tcpDialer
		}
		if overrides.tlsDialer != nil {
			tlsDialer = overrides.tlsDialer
		}
	}
	var tcpCfg *tcptls.Config
	if matrix.NeedsTCP {
		tcpCfg, err = tcptls.NewConfig(tcptls.TCPOptions{
			Address:   addr,
			Dialer:    tcpDialer,
			TLSDialer: tlsDialer,
			Observer:  observer,
			MaxConcurrentDials: func() int {
				if options.MaxConcurrentDials == nil {
					return 0
				}
				return *options.MaxConcurrentDials
			}(),
			WarmBackoffInitial: options.WarmBackoffInitial.Build(),
			WarmBackoffMax:     options.WarmBackoffMax.Build(),
		})
		if err != nil {
			return nil, err
		}
	}

	var newQUICBackend func() carrier.QuicBackend
	if matrix.NeedsQUIC {
		if hasOverrides && overrides.newQUICBackend != nil {
			newQUICBackend = overrides.newQUICBackend
		} else {
			quicCfg := quicBackendOptions{
				context:           ctx,
				address:           addr,
				serverName:        options.Server,
				tlsConfig:         tlsConfig,
				quicOptions:       options.QUICOptions,
				dialer:            outboundDialer,
				congestionControl: congestionControl,
				observer:          observer,
			}
			newQUICBackend = func() carrier.QuicBackend { return newQuicBackend(quicCfg) }
		}
	}

	newBundle := func() (*corebundle.CarrierBundle, error) {
		bundleCfg := corebundle.BundleOptions{
			TCP:            tcpCfg,
			Credentials:    credentials,
			ALPN:           alpn,
			Observer:       observer,
			PoolSize:       matrix.Pool,
			PrewarmOnStart: options.PrewarmOnStart,
			Up:             matrixCarrier(matrix.Up),
			Down:           matrixCarrier(matrix.Down),
		}
		if newQUICBackend != nil {
			bundleCfg.QUIC = newQUICBackend()
		}
		return corebundle.NewCarrierBundle(bundleCfg)
	}

	b, err := newBundle()
	if err != nil {
		return nil, err
	}

	networks := []string{N.NetworkTCP, N.NetworkUDP}
	return &Outbound{
		Adapter:   outbound.NewAdapterWithDialerOptions(C.TypeNowhere, tag, networks, options.DialerOptions),
		logger:    logger,
		observer:  observer,
		server:    server,
		matrix:    matrix,
		bundle:    b,
		newBundle: newBundle,
	}, nil
}

func (o *Outbound) getBundle() (*corebundle.CarrierBundle, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return nil, net.ErrClosed
	}
	if o.bundle != nil {
		return o.bundle, nil
	}
	if o.newBundle == nil {
		return nil, net.ErrClosed
	}
	b, err := o.newBundle()
	if err != nil {
		return nil, err
	}
	o.bundle = b
	return b, nil
}

func (o *Outbound) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		b, err := o.getBundle()
		if err != nil {
			return nil, err
		}
		target, err := targetFromSocksaddr(destination)
		if err != nil {
			return nil, err
		}
		return b.OpenTCP(ctx, target)
	case N.NetworkUDP:
		conn, err := o.ListenPacket(ctx, destination)
		if err != nil {
			return nil, err
		}
		return bufio.NewBindPacketConn(conn, destination), nil
	default:
		return nil, E.Extend(N.ErrUnknownNetwork, network)
	}
}

func (o *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	b, err := o.getBundle()
	if err != nil {
		return nil, err
	}
	target, err := targetFromSocksaddr(destination)
	if err != nil {
		return nil, err
	}
	return b.OpenUDPAsync(ctx, target)
}

func (o *Outbound) InterfaceUpdated() {
	o.mu.Lock()
	bundle := o.bundle
	o.bundle = nil
	o.mu.Unlock()
	if bundle != nil {
		if err := bundle.Close(); err != nil {
			diagnostic.Emit(context.Background(), o.observer, diagnostic.Event{
				Level: diagnostic.LevelWarn, Code: "bundle_close_failed", Component: "sing-box",
				Carrier: diagnostic.CarrierQUIC, Err: err,
			})
		}
	}
}

func (o *Outbound) Close() error {
	o.mu.Lock()
	if o.closed {
		o.mu.Unlock()
		return nil
	}
	o.closed = true
	bundle := o.bundle
	o.bundle = nil
	o.newBundle = nil
	o.mu.Unlock()
	if bundle != nil {
		return bundle.Close()
	}
	return nil
}

func (o *Outbound) MatrixForTest() Matrix { return o.matrix }

func (o *Outbound) bundleForTest() *corebundle.CarrierBundle {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.bundle
}

// socksaddrDialer adapts SingBox N.Dialer (Socksaddr) to tcptls.TCPDialer (host:port string).
type socksaddrDialer struct {
	dialer N.Dialer
}

func (d *socksaddrDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.dialer.DialContext(ctx, network, M.ParseSocksaddr(address))
}

// singTLSDialer adapts sing-box TLS ClientHandshake to tcptls.TlsDialer.
type singTLSDialer struct {
	config tls.Config
}

func (d *singTLSDialer) DialTLSConn(ctx context.Context, c net.Conn) (wire.HandshakedConn, error) {
	conn, err := tls.ClientHandshake(ctx, c, d.config)
	if err != nil {
		return wire.HandshakedConn{}, err
	}
	state := conn.ConnectionState()
	material, err := state.ExportKeyingMaterial(wire.TLSExporterLabel, wire.EmptyTLSExporterContext(), wire.TLSExporterLen)
	if err != nil {
		_ = conn.Close()
		return wire.HandshakedConn{}, err
	}
	var value wire.TLSExporter
	copy(value[:], material)
	return wire.HandshakedConn{
		Conn: conn,
		TLSHandshakeInfo: wire.TLSHandshakeInfo{
			TLSVersion: state.Version, NegotiatedALPN: state.NegotiatedProtocol, Exporter: value,
		},
	}, nil
}

func targetFromSocksaddr(destination M.Socksaddr) (wire.Target, error) {
	if destination.Fqdn != "" {
		return wire.NewDomainTarget(destination.Fqdn, destination.Port)
	}
	return wire.NewIPTarget(destination.Addr, destination.Port)
}

var _ tcptls.TCPDialer = (*socksaddrDialer)(nil)
var _ tcptls.TLSDialer = (*singTLSDialer)(nil)

func matrixCarrier(value string) wire.Carrier {
	if value == "tcp" {
		return wire.CarrierTLSTCP
	}
	return wire.CarrierQUIC
}

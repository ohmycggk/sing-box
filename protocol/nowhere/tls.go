package nowhere

import (
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/common/tls"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json/badoption"

	"github.com/ohmycggk/nowhere-go/wire"
)

const defaultALPN = "now/1"

func normalizeNowhereInboundTLS(options *option.InboundTLSOptions) (*option.InboundTLSOptions, error) {
	if options == nil {
		return nil, C.ErrTLSRequired
	}
	alpn, err := normalizeNowhereALPN(options.ALPN)
	if err != nil {
		return nil, err
	}
	clone := *options
	clone.ALPN = badoption.Listable[string]{alpn}
	clone.MinVersion = "1.3"
	clone.MaxVersion = "1.3"
	return &clone, nil
}

func normalizeNowhereOutboundTLS(options *option.OutboundTLSOptions) (*option.OutboundTLSOptions, error) {
	if options == nil {
		return nil, C.ErrTLSRequired
	}
	alpn, err := normalizeNowhereALPN(options.ALPN)
	if err != nil {
		return nil, err
	}
	clone := *options
	clone.ALPN = badoption.Listable[string]{alpn}
	clone.MinVersion = "1.3"
	clone.MaxVersion = "1.3"
	return &clone, nil
}

// applyNowhereCertificatePin installs a leaf-certificate SHA-256 pin verifier on
// the host TLS client config. Pin overrides SNI/chain checks.
func applyNowhereCertificatePin(config tls.Config, pin string) error {
	if pin == "" {
		return nil
	}
	verifier, err := wire.PeerCertificatePinVerifier(pin)
	if err != nil {
		return err
	}
	std, err := config.STDConfig()
	if err != nil {
		return E.Cause(err, "nowhere: pin requires a std TLS config")
	}
	std.InsecureSkipVerify = true
	std.VerifyPeerCertificate = verifier
	return nil
}

func normalizeNowhereALPN(values []string) (string, error) {
	if len(values) == 0 || len(values) == 1 && values[0] == "" {
		return defaultALPN, nil
	}
	if len(values) != 1 {
		return "", E.New("nowhere: exactly one ALPN is required")
	}
	return values[0], nil
}

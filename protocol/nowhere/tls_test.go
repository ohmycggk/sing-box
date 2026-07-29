package nowhere

import (
	"testing"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/stretchr/testify/require"
)

func TestNormalizeNowhereTLSEmptyALPNAndVersion(t *testing.T) {
	original := &option.InboundTLSOptions{Enabled: true, MinVersion: "1.2", MaxVersion: "1.2"}
	normalized, err := normalizeNowhereInboundTLS(original)
	require.NoError(t, err)
	require.Equal(t, []string{defaultALPN}, []string(normalized.ALPN))
	require.Equal(t, "1.3", normalized.MinVersion)
	require.Equal(t, "1.3", normalized.MaxVersion)
	require.Empty(t, original.ALPN)
	require.Equal(t, "1.2", original.MinVersion)
}

func TestNormalizeNowhereTLSPreservesSingleCustomALPN(t *testing.T) {
	original := &option.OutboundTLSOptions{
		Enabled: true,
		ALPN:    badoption.Listable[string]{"custom/1"},
	}
	normalized, err := normalizeNowhereOutboundTLS(original)
	require.NoError(t, err)
	require.Equal(t, []string{"custom/1"}, []string(normalized.ALPN))
	require.Equal(t, "1.3", normalized.MinVersion)
	require.Equal(t, "1.3", normalized.MaxVersion)
}

func TestNormalizeNowhereTLSRejectsMultipleALPN(t *testing.T) {
	_, err := normalizeNowhereOutboundTLS(&option.OutboundTLSOptions{
		Enabled: true,
		ALPN:    badoption.Listable[string]{"now/1", "h3"},
	})
	require.ErrorContains(t, err, "exactly one ALPN")
}

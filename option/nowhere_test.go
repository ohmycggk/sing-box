package option

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNowhereInboundTLSIsNativeOnly(t *testing.T) {
	t.Parallel()

	t.Run("native TLS object", func(t *testing.T) {
		var options NowhereInboundOptions
		require.NoError(t, json.Unmarshal([]byte(`{
			"password":"secret",
			"tls":{"enabled":true,"alpn":["now/1"],"certificate_path":"c.pem","key_path":"k.pem"}
		}`), &options))
		require.NotNil(t, options.TLS)
		require.True(t, options.TLS.Enabled)
		require.Equal(t, "c.pem", options.TLS.CertificatePath)
		require.Equal(t, "k.pem", options.TLS.KeyPath)
	})

	for _, legacyTLS := range []string{"0", "1", "2", `"1"`, "false", "true"} {
		t.Run("reject legacy tls "+legacyTLS, func(t *testing.T) {
			var options NowhereInboundOptions
			require.Error(t, json.Unmarshal([]byte(`{"password":"secret","tls":`+legacyTLS+`}`), &options))
		})
	}
}

func TestNowhereInboundAdmissionLimitsJSON(t *testing.T) {
	t.Parallel()
	var options NowhereInboundOptions
	require.NoError(t, json.Unmarshal([]byte(`{
		"password":"secret",
		"tls":{"enabled":true,"insecure":true},
		"max_unauthenticated_connections":128,
		"max_unauthenticated_per_source":16
	}`), &options))
	require.NotNil(t, options.MaxUnauthenticatedConnections)
	require.Equal(t, 128, *options.MaxUnauthenticatedConnections)
	require.NotNil(t, options.MaxUnauthenticatedPerSource)
	require.Equal(t, 16, *options.MaxUnauthenticatedPerSource)
}

func TestNowhereOutboundStormControlsJSON(t *testing.T) {
	t.Parallel()
	var options NowhereOutboundOptions
	require.NoError(t, json.Unmarshal([]byte(`{
		"server":"127.0.0.1",
		"server_port":2077,
		"password":"secret",
		"up":"tcp",
		"down":"tcp",
		"max_concurrent_dials":16,
		"warm_backoff_initial":"2s",
		"warm_backoff_max":"20s",
		"tls":{"enabled":true,"insecure":true}
	}`), &options))
	require.NotNil(t, options.MaxConcurrentDials)
	require.Equal(t, 16, *options.MaxConcurrentDials)
	require.Equal(t, 2*time.Second, options.WarmBackoffInitial.Build())
	require.Equal(t, 20*time.Second, options.WarmBackoffMax.Build())
}

func TestNowhereInboundQUICCongestionControlAndRemovedRateFields(t *testing.T) {
	t.Parallel()
	var options NowhereInboundOptions
	require.NoError(t, json.Unmarshal([]byte(`{
		"password":"secret",
		"quic_congestion_control":"bbr2_variant",
		"up_mbps":100,
		"down_mbps":200
	}`), &options))
	require.Equal(t, "bbr2_variant", options.QUICCongestionControl)

	encoded, err := json.Marshal(options)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"quic_congestion_control":"bbr2_variant"`)
	require.NotContains(t, string(encoded), "up_mbps")
	require.NotContains(t, string(encoded), "down_mbps")
}

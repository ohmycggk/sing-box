package nowhere

import "testing"

func TestResolveMatrixDefaultsAndPool(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		up, down    string
		pool        *int
		wantUp      string
		wantDown    string
		wantPool    int
		wantAsym    bool
		wantQUIC    bool
		wantTCP     bool
		wantWarning bool
		wantErr     bool
	}{
		{name: "default-udp-udp", wantUp: "udp", wantDown: "udp", wantPool: 0, wantQUIC: true},
		{name: "tcp-tcp-default-pool", up: "tcp", down: "tcp", wantUp: "tcp", wantDown: "tcp", wantPool: 5, wantTCP: true},
		{name: "tcp-tcp-explicit-pool", up: "tcp", down: "tcp", pool: intPtr(3), wantUp: "tcp", wantDown: "tcp", wantPool: 3, wantTCP: true},
		{name: "tcp-tcp-large-valid-pool", up: "tcp", down: "tcp", pool: intPtr(99), wantUp: "tcp", wantDown: "tcp", wantPool: 99, wantTCP: true},
		{name: "tcp-tcp-over-limit-clamped", up: "tcp", down: "tcp", pool: intPtr(257), wantUp: "tcp", wantDown: "tcp", wantPool: 256, wantTCP: true, wantWarning: true},
		{name: "udp-udp-forces-pool-zero", up: "udp", down: "udp", pool: intPtr(5), wantUp: "udp", wantDown: "udp", wantPool: 0, wantQUIC: true, wantWarning: true},
		{name: "udp-udp-explicit-zero", up: "udp", down: "udp", pool: intPtr(0), wantUp: "udp", wantDown: "udp", wantPool: 0, wantQUIC: true},
		{name: "tcp-udp-forces-pool-zero", up: "tcp", down: "udp", pool: intPtr(5), wantUp: "tcp", wantDown: "udp", wantPool: 0, wantAsym: true, wantQUIC: true, wantTCP: true, wantWarning: true},
		{name: "udp-tcp-forces-pool-zero", up: "udp", down: "tcp", pool: intPtr(5), wantUp: "udp", wantDown: "tcp", wantPool: 0, wantAsym: true, wantQUIC: true, wantTCP: true, wantWarning: true},
		{name: "one-sided-up", up: "tcp", wantErr: true},
		{name: "one-sided-down", down: "udp", wantErr: true},
		{name: "bad-carrier", up: "quic", down: "tcp", wantErr: true},
		{name: "negative-pool", up: "tcp", down: "tcp", pool: intPtr(-1), wantErr: true},
		{name: "negative-pool-udp-ignored", up: "udp", down: "udp", pool: intPtr(-1), wantUp: "udp", wantDown: "udp", wantPool: 0, wantQUIC: true, wantWarning: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveMatrix(tc.up, tc.down, tc.pool)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveMatrix: %v", err)
			}
			if got.Up != tc.wantUp || got.Down != tc.wantDown {
				t.Fatalf("up/down = %s/%s, want %s/%s", got.Up, got.Down, tc.wantUp, tc.wantDown)
			}
			if got.Pool != tc.wantPool {
				t.Fatalf("pool = %d, want %d", got.Pool, tc.wantPool)
			}
			if got.Asymmetric != tc.wantAsym {
				t.Fatalf("asymmetric = %v, want %v", got.Asymmetric, tc.wantAsym)
			}
			if got.NeedsQUIC != tc.wantQUIC || got.NeedsTCP != tc.wantTCP {
				t.Fatalf("needs quic/tcp = %v/%v, want %v/%v", got.NeedsQUIC, got.NeedsTCP, tc.wantQUIC, tc.wantTCP)
			}
			if (got.poolWarning != "") != tc.wantWarning {
				t.Fatalf("pool warning = %q, want warning=%v", got.poolWarning, tc.wantWarning)
			}
			if got.RequiresFlowEnvelope() != tc.wantAsym {
				t.Fatalf("RequiresFlowEnvelope = %v, want %v", got.RequiresFlowEnvelope(), tc.wantAsym)
			}
		})
	}
}

func intPtr(v int) *int { return &v }

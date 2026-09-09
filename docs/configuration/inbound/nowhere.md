---
icon: material/new-box
---

### Structure

```json
{
  "type": "nowhere",
  "tag": "nw-in",

  ... // Listen Fields

  "password": "secret",
  "network": [
    "tcp",
    "udp"
  ],
  "tls": {},
  "quic_congestion_control": "bbr",
  "next": {
    "server": "origin.example",
    "server_port": 2080,
    "password": "origin-key",
    "up": "tcp",
    "down": "tcp",
    "pool": 5,
    "server_name": "origin.example",
    "pin": "<leaf cert sha256 hex>"
  },

  ... // QUIC Fields
}
```

Nowhere Portal is the SingBox inbound for the Nowhere protocol. It accepts
TLS/TCP and (with `with_quic`) QUIC/UDP carriers, authenticates clients, and
routes TCP/UDP targets through the SingBox router (unless chained to another
Portal via `next`).

Plaintext inbound is not allowed. TLS must be enabled.

Nowhere 1.5 is a lockstep upgrade. It binds authentication to the TLS exporter,
so this Portal and every client must run matching 1.5 implementations.

Nowhere 1.7 keeps the 1.5/1.6 authentication and data-plane formats but assigns
the FLOW header's high three bits to a HOPS forwarding budget, enabling native
Portal chaining via `next`. Direct clients still interoperate with 1.5/1.6
Portals; every Portal in a native chain must run Nowhere 1.7 or later, because
older endpoints reject nonzero HOPS as reserved bits.

The bundled nowhere-go dependency is v1.8.3. After AuthFrame, inbound TLS
auto-detects dedicated lanes versus marked Mux shards. Nowhere 1.8.3 `mix` on
`next` is a client policy resolved per flow before FlowHeader.

### Listen Fields

See [Listen Fields](/configuration/shared/listen/) for details.

### Fields

#### password

==Required==

Shared key used to derive auth material. Must match the outbound `password`.

#### network

Listen transports. Default: `["tcp", "udp"]`.

| Value | Behavior |
| --- | --- |
| `["tcp"]` | TLS/TCP listener (TCP relay + UoT). |
| `["udp"]` | QUIC listener (streams + DATAGRAM). Requires build tag `with_quic`. |
| `["tcp", "udp"]` | Both. |

Without `with_quic`, any configuration that enables UDP returns
`QUIC is not included in this build`.

#### max_unauthenticated_connections

Process-wide limit for connections that have not finished authentication
(including the TLS handshake stage). Omitted or `0` defaults to `256`, matching
the Rust Portal.

#### max_unauthenticated_per_source

Per-source unauthenticated connection limit. IPv4 uses `/32`; IPv6 is grouped
by `/64`. Omitted or `0` defaults to `32`. Must not exceed the global limit.

#### tls

==Required==

Use the native SingBox TLS object. See [TLS](/configuration/shared/tls/#inbound).

Nowhere always uses TLS 1.3. If ALPN is omitted, it is normalized to `now/1`;
otherwise exactly one ALPN value is required.

Recommended native example:

```json
{
  "enabled": true,
  "alpn": ["now/1"],
  "min_version": "1.3",
  "max_version": "1.3",
  "certificate_path": "/path/cert.pem",
  "key_path": "/path/key.pem"
}
```

#### quic_congestion_control

QUIC congestion controller used by the inbound sender. Default: `bbr`.

Available values: `bbr`, `bbr_standard`, `bbr2`, `bbr2_variant`, `cubic`, and
`reno`. An unknown value is rejected while constructing the inbound. This
setting only affects QUIC/UDP carriers and does not implement a bandwidth rate
limit. The inbound and outbound may use different controllers.

#### next

Optional native Portal chaining. When set, the Portal forwards **every**
accepted flow to the next Nowhere Portal instead of dialing the target itself:
SingBox routing and detour are bypassed entirely, with no fallback. When
omitted, flows are routed normally.

| Field | Default | Description |
| --- | --- | --- |
| `server` | ==Required== | Next Portal address. |
| `server_port` | ==Required== | Next Portal port. |
| `password` | ==Required== | Shared key of the next Portal. |
| `up`, `down` | `udp` / `udp` | Carrier selectors toward the next Portal (`tcp`, `udp`, or `mix`); both must be set, or both omitted. `mix` is a Nowhere 1.8.3 client policy. Any matrix that can select `udp`/`mix` requires build tag `with_quic`. |
| `mux` | `0` | `0` dedicated TLS lanes, `1` TLS Mux when TCP is possible. `udp/udp&mux=1` canonicalizes to `0`. |
| `pool` | `5` for dedicated `tcp` / `tcp`, otherwise `0` | TLS/TCP warm pool toward the next Portal. Values above `256` are clamped to `256`; `mux=1` and QUIC-capable matrices ignore `pool`. |
| `mix_fallback_timeout` | `1s` | Primary mix-route preparation budget. |
| `server_name` | none | DNS name used to verify the next Portal's TLS certificate. Omitted, empty, or the literal `"none"` disables certificate verification; the endpoint host may still be sent as the ClientHello SNI. |
| `pin` | none | SHA-256 hex fingerprint of the next Portal's leaf certificate. Omitted, empty, or `"none"` disables pinning; a real pin overrides `server_name` and chain verification. |

The chained upstream client inherits the inbound TLS ALPN.

Hop budget: the first forwarding Portal initializes HOPS to 7, each
Portal-to-Portal hop decrements it by one, and a Portal that would forward an
exhausted flow (HOPS=1) rejects it with `FLOW_LIMIT` before opening the next
hop. Setup rejection codes from an upstream Portal propagate downstream
unchanged. Every Portal in a chain must run Nowhere 1.7 or later.

### QUIC Fields

See [QUIC](/configuration/shared/quic/) for shared QUIC options when present.
The same window, idle timeout, keepalive, stream-limit, packet-size, and PMTU
settings are applied by both Nowhere inbound and outbound.

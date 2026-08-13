---
icon: material/new-box
---

### Structure

```json
{
  "type": "nowhere",
  "tag": "nw-out",

  "server": "example.com",
  "server_port": 2077,
  "password": "secret",
  "up": "udp",
  "down": "udp",
  "pool": 0,
  "tls": {},
  "quic_congestion_control": "bbr",

  ... // Dial Fields
  ... // QUIC Fields
}
```

Nowhere outbound is the SingBox client for the Nowhere protocol. Upload and
download carriers are selected independently via `up` / `down`.

Nowhere 1.5 is a lockstep upgrade. It binds authentication to the TLS exporter,
so the Portal and every client must run matching 1.5 implementations.

Nowhere 1.7 keeps the 1.5/1.6 authentication and direct-flow data plane. This
outbound sends HOPS=0, so direct connections remain compatible with 1.5/1.6
Portals. Non-zero HOPS is emitted only by a 1.7 native forwarding Portal.

### Fields

#### server

==Required==

The server address.

#### server_port

==Required==

The server port.

#### password

==Required==

Shared key. Must match the inbound `password`.

#### up, down

Carrier selectors: `"tcp"` or `"udp"`.

Both must be set, or both omitted (default `udp` / `udp`).

| Matrix | TCP traffic | UDP traffic | Notes |
| --- | --- | --- | --- |
| `tcp` / `tcp` | One TLS connection | UoT on TLS | `pool` applies (default `5`, max `256`). |
| `udp` / `udp` | One QUIC stream | QUIC DATAGRAM | Requires `with_quic`. |
| `tcp` / `udp` | TLS upload + QUIC download | UoT upload + DATAGRAM download | Asymmetric; requires `with_quic`. |
| `udp` / `tcp` | QUIC upload + TLS download | DATAGRAM upload + UoT download | Asymmetric; requires `with_quic`. |

Any matrix that includes UDP forces `pool=0`. A non-zero configured value is
normalized to zero and reported once as a configuration warning.

Without `with_quic`, only `tcp` / `tcp` can be constructed; other matrices
return `QUIC is not included in this build`.

#### pool

TLS/TCP warm pool size. Effective only for `tcp` / `tcp`.

Default `5` when omitted on `tcp` / `tcp`. Negative values are rejected and
values above `256` are clamped to `256`. Matrices containing UDP ignore `pool`
entirely, matching the Rust v1.7 parser.

`pool=0` only disables warm preconnect; it does **not** cap business fresh
dials. On `tcp` / `tcp`, each user TCP/UoT flow still consumes one dedicated
TLS/TCP carrier.

#### max_concurrent_dials

Maximum in-flight physical TLS/TCP dials per outbound (dial + TLS + auth).
Omitted or `0` defaults to `16`. Negative values are rejected.

The cap applies to both business fresh dials and warm prepare. Warm prepare
only takes an idle slot non-blockingly and never preempts business dials.

#### warm_backoff_initial, warm_backoff_max

Exponential backoff window after warm-prepare failures. Defaults are `1s` /
`30s` when omitted. The initial value must not exceed the max. Warm replenish
is skipped during the backoff window; the pool only retries after a successful
business fresh dial.

#### tls

==Required==

TLS configuration, see [TLS](/configuration/shared/tls/#outbound).

Must enable TLS. Nowhere forces TLS 1.3. If ALPN is omitted, it is normalized
to `now/1`; otherwise exactly one ALPN value is required.

For a self-signed server, set `insecure: true` or pin the certificate /
fingerprint explicitly.

#### quic_congestion_control

QUIC congestion controller used by the outbound sender. Default: `bbr`.

Available values: `bbr`, `bbr_standard`, `bbr2`, `bbr2_variant`, `cubic`, and
`reno`. An unknown value is rejected while constructing the outbound, including
TCP-only carrier configurations. This setting only affects QUIC/UDP carriers
and does not implement a bandwidth rate limit. The inbound and outbound may use
different controllers.

### Dial Fields

See [Dial Fields](/configuration/shared/dial/) for details.

### QUIC Fields

See [QUIC](/configuration/shared/quic/) for shared QUIC options when present.
The same window, idle timeout, keepalive, stream-limit, packet-size, and PMTU
settings are applied by both Nowhere inbound and outbound.

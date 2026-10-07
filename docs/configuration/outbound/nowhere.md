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
  "password": "0123456789abcdef0123456789abcdef",
  "up": "udp",
  "down": "udp",
  "morph": false,
  "pool": 0,
  "tls": {},
  "quic_congestion_control": "bbr",

  ... // Dial Fields
  ... // QUIC Fields
}
```

Nowhere outbound is the SingBox client for the Nowhere protocol. Upload and
download carriers are selected independently via `up` / `down`.

This is a Nowhere 2.2.1 (`nw2`) implementation. It negotiates the `nw2` ALPN
only and does not interoperate with legacy 1.8 peers. Authentication is an
exporter-bound AuthFrame, the FlowHeader is 5 bytes and carries a HOPS
forwarding budget of 0..7, TLS Mux is credit-windowed (`mux=0` dedicated
lanes, `mux=1` marked shards), and QUIC UDP headers are packed.

`morph` enables the Nowhere 2 Morph keyed transform for every carrier of this
endpoint. Morph is not negotiated in-band: `morph=1` must be enabled on both
ends of a hop, because the Nowhere 2.2.1 Morph transform is wire-incompatible
with the 2.0.x Morph.

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

As a client this endpoint is lenient: the key accepts 1–255
URL-percent-decoded bytes and applies no character rule. The Portal side is
stricter — 32–64 lowercase hexadecimal characters — see the
[Nowhere inbound](/configuration/inbound/nowhere/) `password` field.

#### up, down

Carrier selectors: `"tcp"` or `"udp"`.

Both must be set, or both omitted (default `udp` / `udp`).

| Matrix | TCP traffic | UDP traffic | Notes |
| --- | --- | --- | --- |
| `tcp` / `tcp` | One TLS connection | UoT on TLS | `pool` applies for `mux=0` (default `5`, max `256`). |
| `udp` / `udp` | One QUIC stream | QUIC DATAGRAM | Requires `with_quic`. `mux=1` canonicalizes to `0`. |
| `tcp` / `udp` | TLS upload + QUIC download | UoT upload + DATAGRAM download | Asymmetric; requires `with_quic`. |
| `udp` / `tcp` | QUIC upload + TLS download | DATAGRAM upload + UoT download | Asymmetric; requires `with_quic`. |

Any matrix that can select QUIC (`udp`) forces `pool=0`. A non-zero
configured value is normalized to zero and reported once as a configuration
warning.

Without `with_quic`, only `tcp` / `tcp` can be constructed; other matrices
return `QUIC is not included in this build`.

#### mux

TLS lane framing. `0` (default) uses dedicated TLS lanes; `1` enables marked Mux
shards. Mux applies when either direction is `tcp`. `udp/udp&mux=1`
canonicalizes to `0`.

#### pool

TLS/TCP warm pool size. Effective only for dedicated (`mux=0`) `tcp` / `tcp`.

Default `5` when omitted on dedicated `tcp` / `tcp`. Negative values are rejected and
values above `256` are clamped to `256`. `mux=1` and any matrix that can select
QUIC ignore `pool`.

`pool=0` only disables warm preconnect; it does **not** cap business fresh
dials. On `tcp` / `tcp`, each user TCP/UoT flow still consumes one dedicated
TLS/TCP carrier.

#### morph

Enable the Nowhere 2 Morph keyed transform for every carrier of this endpoint:
a 64-byte random TCP prelude is sent below TLS before authentication, and
directional ChaCha20 keys derived from the shared key wrap traffic below both
TLS and QUIC. Default: `false`.

The TCP prelude policy is `full8` by default (all eight bits of each prelude
byte are random). Set the `NOW_MORPH_TCP_PRELUDE` environment variable to
`low7` for the older policy that clears each byte's high bit; a set-but-empty
value is rejected. Both ends of a hop must use the same policy.

There is no in-band negotiation. `morph` must be enabled on both ends of a hop
(client and Portal, or Portal and next Portal), and the Nowhere 2.2.1 Morph
transform is wire-incompatible with the 2.0.x Morph, so every peer on a
`morph=1` hop must run Nowhere 2.2.1.

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
to `nw2`; otherwise exactly one ALPN value is required.

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

The shared `inet4_bind_address` / `inet6_bind_address` fields bind the source
address per address family for this outbound. They are the sing-box equivalent
of the Rust `dial` parameter and are mutually exclusive with the Rust `dial4` /
`dial6` pair, which this build exposes only on the inbound `next` hop (see
[Nowhere inbound](/configuration/inbound/nowhere/#next)).

### QUIC Fields

See [QUIC](/configuration/shared/quic/) for shared QUIC options when present.
The same window, idle timeout, keepalive, stream-limit, packet-size, and PMTU
settings are applied by both Nowhere inbound and outbound.

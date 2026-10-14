---
icon: material/new-box
---

### Structure

```json
{
  "type": "nowhere",
  "tag": "nw-in",

  ... // Listen Fields

  "password": "0123456789abcdef0123456789abcdef",
  "morph": false,
  "network": [
    "tcp",
    "udp"
  ],
  "tls": {},
  "quic_congestion_control": "bbr",
  "next": {
    "server": "origin.example",
    "server_port": 2080,
    "password": "0123456789abcdef0123456789abcde0",
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

This is a Nowhere 2.2.1 (`nw2`) implementation. It negotiates the `nw2` ALPN
only and does not interoperate with legacy 1.8 peers. Authentication is an
exporter-bound AuthFrame, the FlowHeader is 5 bytes and carries a HOPS
forwarding budget of 0..7, and QUIC UDP headers are packed. After AuthFrame,
inbound TLS auto-detects dedicated lanes (`mux=0`) versus credit-windowed
marked Mux shards (`mux=1`).

Set `"version": 3` to run the Nowhere 3 (`nw3`) stack instead — see
[Nowhere 3](#nowhere-3-nw3) below. The two stacks share no handshake: a single
inbound runs exactly one version, and omitting `version` (or `"version": 2`)
keeps every existing `nw2` configuration working unchanged.

`morph` enables the Nowhere 2 Morph keyed transform for every carrier of this
endpoint. Morph is not negotiated in-band: `morph=1` must be enabled on both
ends of a hop, because the Nowhere 2.2.1 Morph transform is wire-incompatible
with the 2.0.x Morph.

### Listen Fields

See [Listen Fields](/configuration/shared/listen/) for details.

### Fields

#### password

==Required==

Shared key used to derive auth material. Must match the outbound `password`.

As a Portal this endpoint applies the Nowhere 2.2.1 key admission rule: after
URL percent-decoding the key must be 32–64 lowercase hexadecimal characters
(`[0-9a-f]{32,64}`, odd lengths accepted). The key text is used directly as the
HKDF input; it is never hex-decoded. Generate one with
`nowhere generate-key`. Client-side (outbound) keys stay lenient at 1–255
decoded bytes.

#### morph

Enable the Nowhere 2 Morph keyed transform for every carrier of this Portal:
a 64-byte random TCP prelude is consumed below TLS before authentication, and
directional ChaCha20 keys derived from the shared key unwrap traffic below
both TLS and QUIC. Default: `false`.

The TCP prelude policy is `full8` by default (all eight bits of each prelude
byte are random). Set the `NOW_MORPH_TCP_PRELUDE` environment variable to
`low7` for the older policy that clears each byte's high bit; a set-but-empty
value is rejected. Both ends of a hop must use the same policy.

There is no in-band negotiation: `morph` must match the client. For a chained
Portal, each hop inherits this setting unless `next.morph` overrides it. The
Nowhere 2.2.1 Morph transform is wire-incompatible with the 2.0.x Morph, so
every peer on a `morph=1` hop must run Nowhere 2.2.1.

#### network

Listen transports. Default: `["tcp", "udp"]`.

| Value | Behavior |
| --- | --- |
| `["tcp"]` | TLS/TCP listener (TCP relay + UoT). |
| `["udp"]` | QUIC listener (streams + DATAGRAM). Requires build tag `with_quic`. |
| `["tcp", "udp"]` | Both. |

Without `with_quic`, any configuration that enables UDP returns
`QUIC is not included in this build`.

Under `"version": 3` the carrier is selected by `carrier` instead, and an
explicit `network` may only name that one transport (see
[Nowhere 3](#nowhere-3-nw3)).

#### version

Nowhere protocol stack. `2` (default) is Nowhere 2.2.1 (`nw2`); `3` is
Nowhere 3 (`nw3`). Omitted or `0` means `2`.

The stacks share no handshake, so a single inbound runs exactly one version.
Any other value is rejected at construction time.

#### carrier

Nowhere 3 transport: `"tcp"` (TLS over TCP, UDP flows carried as UoT) or
`"quic"` (QUIC/UDP, streams + DATAGRAM). Empty means `"tcp"`.

Only meaningful with `"version": 3`; the `nw2` stack selects its carriers via
`network` (inbound) and `up` / `down` (outbound, chained `next`) instead.

`"quic"` requires the build tag `with_quic`.

#### egress

Nowhere 3 exit-bridge mode. When `true`, the Portal does not route flows
locally: it accepts a single
[Nowhere Egress](/configuration/inbound/nowhere-egress/) registration over
QUIC and forwards every authenticated flow to that exit.

Requires `"version": 3` and `"carrier": "quic"` — the registration handshake
only exists on the QUIC carrier.

### Nowhere 3 (nw3)

`"version": 3` selects the Nowhere 3 stack, a ground-up rewrite that shares no
wire format with `nw2`. Differences from the fields above:

| Aspect | Nowhere 3 behavior |
| --- | --- |
| ALPN | Pinned to `nw3`. `tls.alpn` is rejected. |
| TLS version | Pinned to TLS 1.3. `tls.min_version` / `tls.max_version` are rejected. |
| TLS stack | Standard library only. `tls.utls` / `tls.ech` / `tls.reality` are rejected; inbound `tls.acme` / `tls.certificate_provider` are rejected (a static certificate is required). |
| Key | Exactly 32 lowercase hex characters (`[0-9a-f]{32}`). The lenient `nw2` rules (32–64, odd lengths) do not apply. |
| Morph | Defaults to **on**. An explicit `morph` still wins. The Nowhere 3 Morph transform uses a different derivation than `nw2`, so both ends of a hop must run the same version. |
| Carriers | Exactly one per endpoint, selected by `carrier`. |
| Chaining | `next` is rejected; Nowhere 3 has no native Portal chaining. |
| Admission limits | `max_unauthenticated_connections` / `max_unauthenticated_per_source` are rejected; the vendored Portal owns its own limits. |
| Networks | An explicit `network` must equal the carrier's transport: `tcp` carrier implies `["tcp"]`, `quic` carrier implies `["udp"]`. |

A minimal Nowhere 3 Portal over the TCP carrier:

```json
{
  "type": "nowhere",
  "tag": "nw3-in",
  "listen_port": 2080,
  "password": "0f1e2d3c4b5a69788796a5b4c3d2e1f0",
  "version": 3,
  "carrier": "tcp",
  "tls": {
    "enabled": true,
    "certificate_path": "/path/cert.pem",
    "key_path": "/path/key.pem"
  }
}
```

On startup the Portal logs the leaf certificate's SHA-256 fingerprint so a
`pin=` value can be obtained for the matching
[outbound](/configuration/outbound/nowhere/#nowhere-3-nw3).

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

Nowhere always uses TLS 1.3. If ALPN is omitted, it is normalized to `nw2`;
otherwise exactly one ALPN value is required.

Recommended native example:

```json
{
  "enabled": true,
  "alpn": ["nw2"],
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
| `password` | ==Required== | Shared key of the next Portal. Subject to the same Portal key rule as the inbound `password`: 32–64 lowercase hexadecimal characters after URL percent-decoding. |
| `up`, `down` | `udp` / `udp` | Carrier selectors toward the next Portal (`tcp` or `udp`); both must be set, or both omitted. Any matrix that can select `udp` requires build tag `with_quic`. |
| `mux` | `0` | `0` dedicated TLS lanes, `1` credit-windowed TLS Mux when TCP is possible. `udp/udp&mux=1` canonicalizes to `0`. |
| `pool` | `5` for dedicated `tcp` / `tcp`, otherwise `0` | TLS/TCP warm pool toward the next Portal. Values above `256` are clamped to `256`; `mux=1` and QUIC-capable matrices ignore `pool`. |
| `server_name` | none | DNS name used to verify the next Portal's TLS certificate. Omitted, empty, or the literal `"none"` disables certificate verification; the endpoint host may still be sent as the ClientHello SNI. |
| `pin` | none | SHA-256 hex fingerprint of the next Portal's leaf certificate. Omitted, empty, or `"none"` disables pinning; a real pin overrides `server_name` and chain verification. |
| `morph` | inherit inbound `morph` | Enable the Nowhere 2 Morph keyed transform toward the next Portal. Omitted (`nil`) inherits the inbound `morph`; set `true` / `false` to override the hop. Both ends of the hop must agree. |
| `dial4` | `auto` | IPv4 source address bound when dialing the next Portal over IPv4. Accepts `auto` (leave the source address to the system) or an IPv4 literal; `0.0.0.0` is a valid wildcard. |
| `dial6` | `auto` | IPv6 source address bound when dialing the next Portal over IPv6. Accepts `auto` or an IPv6 literal; IPv4-mapped literals such as `::ffff:192.0.2.1` are rejected and `::` is a valid wildcard. |

`dial4` and `dial6` are independent and may be combined for dual-stack source
binding. They mirror the Rust portal URL parameters `dial4` / `dial6` and are
mutually exclusive with `dial`: this build has no `dial` option for the next
hop, and the equivalent legacy single-family binding on an outbound is the
shared `inet4_bind_address` / `inet6_bind_address` dial field. A configured
bind is never silently downgraded to automatic — a failed bind fails the
dial and the next candidate is tried.

The chained upstream client inherits the inbound TLS ALPN.

Hop budget: the first forwarding Portal initializes HOPS to 7, each
Portal-to-Portal hop decrements it by one, and a Portal that would forward an
exhausted flow (HOPS=1) rejects it with `FLOW_LIMIT` before opening the next
hop. Setup rejection codes from an upstream Portal propagate downstream
unchanged. Every Portal in a chain must run Nowhere 2.2.1 (`nw2`).

### QUIC Fields

See [QUIC](/configuration/shared/quic/) for shared QUIC options when present.
The same window, idle timeout, keepalive, stream-limit, packet-size, and PMTU
settings are applied by both Nowhere inbound and outbound.

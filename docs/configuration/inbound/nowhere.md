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

  ... // QUIC Fields
}
```

Nowhere Portal is the SingBox inbound for the Nowhere protocol. It accepts
TLS/TCP and (with `with_quic`) QUIC/UDP carriers, authenticates clients, and
routes TCP/UDP targets through the SingBox router.

Plaintext inbound is not allowed. TLS must be enabled.

Nowhere 1.5 is a lockstep upgrade. It binds authentication to the TLS exporter,
so this Portal and every client must run matching 1.5 implementations.

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

### QUIC Fields

See [QUIC](/configuration/shared/quic/) for shared QUIC options when present.
The same window, idle timeout, keepalive, stream-limit, packet-size, and PMTU
settings are applied by both Nowhere inbound and outbound.

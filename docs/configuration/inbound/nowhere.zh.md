---
icon: material/new-box
---

### 结构

```json
{
  "type": "nowhere",
  "tag": "nw-in",

  ... // 监听字段

  "password": "secret",
  "network": [
    "tcp",
    "udp"
  ],
  "tls": {},
  "quic_congestion_control": "bbr",

  ... // QUIC 字段
}
```

Nowhere Portal 是 SingBox 的 Nowhere 入站。它接受 TLS/TCP，以及（在
`with_quic` 构建下）QUIC/UDP carrier，完成认证后把 TCP/UDP 目标交给
SingBox 路由。

不允许明文入站，必须启用 TLS。

Nowhere 1.5 必须锁步升级。认证绑定 TLS exporter，本 Portal 与所有客户端必须运行
匹配的 1.5 实现。

### 监听字段

参阅 [监听字段](/zh/configuration/shared/listen/)。

### 字段

#### password

==必填==

用于派生认证材料的共享密钥，必须与出站 `password` 一致。

#### network

监听传输。默认：`["tcp", "udp"]`。

| 值 | 行为 |
| --- | --- |
| `["tcp"]` | TLS/TCP 监听（TCP 中继 + UoT）。 |
| `["udp"]` | QUIC 监听（stream + DATAGRAM）。需要构建标签 `with_quic`。 |
| `["tcp", "udp"]` | 两者都开。 |

无 `with_quic` 时，启用 UDP 的配置会返回
`QUIC is not included in this build`。

#### max_unauthenticated_connections

未完成认证（含 TLS 握手阶段）的全局连接上限。省略或 `0` 时默认 `256`，与
Rust Portal 一致。

#### max_unauthenticated_per_source

每个来源的未认证连接上限。IPv4 按 `/32`，IPv6 按 `/64` 聚合。省略或 `0` 时
默认 `32`。不得超过全局上限。

#### tls

==必填==

使用原生 SingBox TLS 对象。参阅 [TLS](/zh/configuration/shared/tls/#入站)。

Nowhere 固定使用 TLS 1.3。省略 ALPN 时会补全为 `now/1`；显式配置时必须且只能
提供一个 ALPN。

推荐原生示例：

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

入站发送方向使用的 QUIC 拥塞控制器。默认：`bbr`。

可选值：`bbr`、`bbr_standard`、`bbr2`、`bbr2_variant`、`cubic`、`reno`。
未知值会在构造入站时被拒绝。该配置只影响 QUIC/UDP carrier，不提供带宽限速；
入站和出站可以使用不同的控制器。

### QUIC 字段

若存在共享 QUIC 选项，参阅 [QUIC](/zh/configuration/shared/quic/)。
Nowhere 入站和出站会一致应用窗口、idle timeout、keepalive、stream 上限、
初始包大小与 PMTU 配置。

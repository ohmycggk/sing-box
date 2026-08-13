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

  ... // QUIC 字段
}
```

Nowhere Portal 是 SingBox 的 Nowhere 入站。它接受 TLS/TCP，以及（在
`with_quic` 构建下）QUIC/UDP carrier，完成认证后把 TCP/UDP 目标交给
SingBox 路由（除非通过 `next` 链式转发到另一个 Portal）。

不允许明文入站，必须启用 TLS。

Nowhere 1.5 必须锁步升级。认证绑定 TLS exporter，本 Portal 与所有客户端必须运行
匹配的 1.5 实现。

Nowhere 1.7 保留 1.5/1.6 的认证与数据面格式，但将 FLOW 头部的高 3 位用作 HOPS
转发预算，从而支持通过 `next` 进行原生 Portal 链式转发。内置 nowhere-go 依赖已
升级为 v1.7.0。直连客户端仍可与 1.5/1.6 Portal 互通；链中的每个 Portal 必须运行
Nowhere 1.7 或更高版本，因为旧版本端点会将非零 HOPS 视为保留位并拒绝。

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

#### next

可选的原生 Portal 链式转发。配置后，本 Portal 会将**所有**已接受的 flow 转发到
下一个 Nowhere Portal，而不再自行拨号：完全绕过 SingBox 路由与 detour，且没有
回退。省略时按正常路由处理。

| 字段 | 默认值 | 说明 |
| --- | --- | --- |
| `server` | ==必填== | 下一个 Portal 的地址。 |
| `server_port` | ==必填== | 下一个 Portal 的端口。 |
| `password` | ==必填== | 下一个 Portal 的共享密钥。 |
| `up`, `down` | `udp` / `udp` | 通往下一个 Portal 的 carrier 选择；必须同时设置或同时省略。任何包含 `udp` 的矩阵都需要构建标签 `with_quic`。 |
| `pool` | `tcp` / `tcp` 时为 `5`，否则为 `0` | 通往下一个 Portal 的 TLS/TCP 预热连接池。超过 `256` 的值会钳制为 `256`；包含 UDP 的矩阵完全忽略 `pool`。 |
| `server_name` | 无 | 用于校验下一个 Portal TLS 证书的 DNS 名称。省略、为空或字面量 `"none"` 时关闭证书校验；端点主机名仍可能作为 ClientHello SNI 发送。 |
| `pin` | 无 | 下一个 Portal 叶子证书的 SHA-256 十六进制指纹。省略、为空或 `"none"` 时关闭证书固定；设置有效 pin 时优先于 `server_name` 与证书链校验。 |

链式上游客户端继承入站的 TLS ALPN。

跳数预算：第一个转发 Portal 将 HOPS 初始化为 7，每一跳 Portal 间转发将其递减 1；
当需要继续转发预算耗尽的 flow（HOPS=1）时，Portal 会在开启下一跳之前以
`FLOW_LIMIT` 拒绝。上游 Portal 的 setup 拒绝码会原样传回下游。链中的每个
Portal 都必须运行 Nowhere 1.7 或更高版本。

### QUIC 字段

若存在共享 QUIC 选项，参阅 [QUIC](/zh/configuration/shared/quic/)。
Nowhere 入站和出站会一致应用窗口、idle timeout、keepalive、stream 上限、
初始包大小与 PMTU 配置。

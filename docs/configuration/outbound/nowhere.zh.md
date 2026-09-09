---
icon: material/new-box
---

### 结构

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

  ... // 拨号字段
  ... // QUIC 字段
}
```

Nowhere 出站是 SingBox 的 Nowhere 客户端。上传与下载 carrier 通过 `up` /
`down` 独立选择。

Nowhere 1.5 必须锁步升级。认证绑定 TLS exporter，Portal 与所有客户端必须运行
匹配的 1.5 实现。

Nowhere 1.7 保留 1.5/1.6 的认证和直连 flow 数据面。本出站发送 HOPS=0，因此直连
仍兼容 1.5/1.6 Portal；只有 1.7 原生转发 Portal 才会发送非零 HOPS。

Nowhere 1.8 新增 TLS Mux（`mux=0` 专用通道，`mux=1` 标记分片）。`mux=0` 客户端
仍可对接 1.8 Portal。Nowhere 1.8.3 新增仅客户端的 `mix` 策略；数据面与 1.8.2
相同，FlowHeader 仍只携带 TT / TQ / QT / QQ。

### 字段

#### server

==必填==

服务器地址。

#### server_port

==必填==

服务器端口。

#### password

==必填==

共享密钥，必须与入站 `password` 一致。

#### up, down

Carrier 选择：`"tcp"`、`"udp"` 或 `"mix"`。

必须同时设置，或同时省略（默认 `udp` / `udp`）。
`mix` 是 Nowhere 1.8.3 的客户端策略，在写入 FlowHeader 之前按 flow 解析。
`mix/mix` 只会解析为 `tcp/tcp` 或 `udp/udp`；单边 `mix` 可以解析为非对称对。
主路由有 1s 准备预算，超时后用新 flow ID 尝试另一条允许的载体对。

| 矩阵 | TCP 流量 | UDP 流量 | 说明 |
| --- | --- | --- | --- |
| `tcp` / `tcp` | 一条 TLS 连接 | TLS 上的 UoT | `mux=0` 时 `pool` 生效（默认 `5`，最大 `256`）。 |
| `udp` / `udp` | 一条 QUIC stream | QUIC DATAGRAM | 需要 `with_quic`。`mux=1` 会规范为 `0`。 |
| `tcp` / `udp` | TLS 上传 + QUIC 下载 | UoT 上传 + DATAGRAM 下载 | 非对称；需要 `with_quic`。 |
| `udp` / `tcp` | QUIC 上传 + TLS 下载 | DATAGRAM 上传 + UoT 下载 | 非对称；需要 `with_quic`。 |
| `mix` / `mix` | 按 flow 选择 `tcp/tcp` 或 `udp/udp` | 与解析结果相同 | 客户端策略；需要 `with_quic`。 |

任何可能选择 QUIC（`udp` 或 `mix`）的矩阵都会强制 `pool=0`。显式配置非零值时
会归一化为零，并记录一次配置警告。

无 `with_quic` 时只能构造 `tcp` / `tcp`；其它矩阵返回
`QUIC is not included in this build`。

#### mux

TLS 通道成帧。`0`（默认）使用专用 TLS 通道；`1` 启用 Mux 分片。任一方向为 `tcp`
或 `mix` 时生效；`udp/udp&mux=1` 会规范为 `0`。

#### mix_fallback_timeout

mix 主路由准备超时。省略或 `0` 使用 `1s`。

#### pool

TLS/TCP 预热连接池大小。仅对 `mux=0` 的 `tcp` / `tcp` 生效。

在专用 `tcp` / `tcp` 下省略时默认 `5`。负数会被拒绝；超过 `256` 的值会钳制为
`256`。`mux=1` 以及任何可能选择 QUIC 的矩阵都会忽略 `pool`。

`pool=0` 只关闭预热，**不会**限制业务 fresh dial。`tcp` / `tcp` 下每个用户
TCP/UoT flow 仍会消耗一条独立 TLS/TCP carrier。

#### max_concurrent_dials

每个 outbound 同时进行的物理 TLS/TCP 建连上限（含 dial、TLS、auth）。省略或
`0` 时默认 `16`。负数会被拒绝。

该上限同时约束业务 fresh dial 与 warm prepare；warm prepare 只会非阻塞占用空闲
slot，不会抢占业务建连。

#### warm_backoff_initial, warm_backoff_max

warm prepare 失败后的指数退避区间。省略时默认 `1s` / `30s`。初始值不得大于
最大值。失败期间不会在退避窗口内继续补池；业务 fresh 成功后才会再次尝试补池。

#### tls

==必填==

TLS 配置，参阅 [TLS](/zh/configuration/shared/tls/#出站)。

必须启用 TLS。Nowhere 固定使用 TLS 1.3。省略 ALPN 时会补全为 `now/1`；
显式配置时必须且只能提供一个 ALPN。

对接自签服务端时，需设置 `insecure: true`，或显式固定证书 / 指纹。

#### quic_congestion_control

出站发送方向使用的 QUIC 拥塞控制器。默认：`bbr`。

可选值：`bbr`、`bbr_standard`、`bbr2`、`bbr2_variant`、`cubic`、`reno`。
未知值会在构造出站时被拒绝，包括仅使用 TCP carrier 的配置。该配置只影响
QUIC/UDP carrier，不提供带宽限速；入站和出站可以使用不同的控制器。

### 拨号字段

参阅 [拨号字段](/zh/configuration/shared/dial/)。

### QUIC 字段

若存在共享 QUIC 选项，参阅 [QUIC](/zh/configuration/shared/quic/)。
Nowhere 入站和出站会一致应用窗口、idle timeout、keepalive、stream 上限、
初始包大小与 PMTU 配置。

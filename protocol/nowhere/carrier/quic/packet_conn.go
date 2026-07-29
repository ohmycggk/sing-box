//go:build with_quic

package quic

import (
	"net"

	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/bufio"
)

type packetConnBufferControl interface {
	SetReadBuffer(int) error
	SetWriteBuffer(int) error
}

type packetConnWithBufferControl struct {
	net.PacketConn
	control packetConnBufferControl
}

// newQUICPacketConn adapts a connected UDP socket without inspecting QUIC payloads.
func newQUICPacketConn(conn net.Conn) net.PacketConn {
	packetConn := bufio.NewUnbindPacketConn(conn)
	control, loaded := common.Cast[packetConnBufferControl](conn)
	if !loaded {
		return packetConn
	}
	return &packetConnWithBufferControl{
		PacketConn: packetConn,
		control:    control,
	}
}

func (c *packetConnWithBufferControl) SetReadBuffer(size int) error {
	return c.control.SetReadBuffer(size)
}

func (c *packetConnWithBufferControl) SetWriteBuffer(size int) error {
	return c.control.SetWriteBuffer(size)
}

var _ net.PacketConn = (*packetConnWithBufferControl)(nil)

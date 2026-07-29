//go:build with_quic

package quic

import (
	"context"
	"time"

	"github.com/sagernet/quic-go"
	"github.com/sagernet/quic-go/congestion"
	"github.com/sagernet/sing-box/protocol/nowhere/internal/quicsettings"
	"github.com/sagernet/sing-quic/congestion_bbr1"
	"github.com/sagernet/sing-quic/congestion_bbr2"
	congestion_meta1 "github.com/sagernet/sing-quic/congestion_meta1"
	congestion_meta2 "github.com/sagernet/sing-quic/congestion_meta2"
	"github.com/sagernet/sing/common/ntp"
)

type CongestionControl = quicsettings.CongestionControl

const (
	CongestionControlBBR         = quicsettings.CongestionControlBBR
	CongestionControlBBRStandard = quicsettings.CongestionControlBBRStandard
	CongestionControlBBR2        = quicsettings.CongestionControlBBR2
	CongestionControlBBR2Variant = quicsettings.CongestionControlBBR2Variant
	CongestionControlCubic       = quicsettings.CongestionControlCubic
	CongestionControlReno        = quicsettings.CongestionControlReno
)

var ParseCongestionControl = quicsettings.ParseCongestionControl

// NewCongestionController constructs the selected controller for one QUIC connection.
func NewCongestionController(ctx context.Context, connection *quic.Conn, selected CongestionControl) congestion.CongestionControl {
	timeFunc := ntp.TimeFuncFromContext(ctx)
	if timeFunc == nil {
		timeFunc = time.Now
	}
	packetSize := congestion.ByteCount(connection.Config().InitialPacketSize)
	switch selected {
	case CongestionControlBBRStandard:
		return congestion_bbr1.NewBbrSender(
			congestion_bbr1.DefaultClock{TimeFunc: timeFunc},
			packetSize,
			congestion_bbr1.InitialCongestionWindowPackets,
			congestion_bbr1.MaxCongestionWindowPackets,
		)
	case CongestionControlBBR2:
		return congestion_bbr2.NewBBR2Sender(
			congestion_bbr2.DefaultClock{TimeFunc: timeFunc},
			packetSize,
			0,
			false,
		)
	case CongestionControlBBR2Variant:
		return congestion_bbr2.NewBBR2Sender(
			congestion_bbr2.DefaultClock{TimeFunc: timeFunc},
			packetSize,
			32*packetSize,
			true,
		)
	case CongestionControlCubic:
		return congestion_meta1.NewCubicSender(
			congestion_meta1.DefaultClock{TimeFunc: timeFunc},
			packetSize,
			false,
		)
	case CongestionControlReno:
		return congestion_meta1.NewCubicSender(
			congestion_meta1.DefaultClock{TimeFunc: timeFunc},
			packetSize,
			true,
		)
	case CongestionControlBBR:
		fallthrough
	default:
		return congestion_meta2.NewBbrSender(
			congestion_meta2.DefaultClock{TimeFunc: timeFunc},
			packetSize,
			congestion.ByteCount(congestion_meta1.InitialCongestionWindow),
		)
	}
}

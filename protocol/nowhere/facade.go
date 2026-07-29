package nowhere

import (
	"github.com/ohmycggk/nowhere-go/carrier/tcptls"
	"github.com/ohmycggk/nowhere-go/wire"
)

type (
	Credentials = wire.Credentials
	SessionID   = wire.SessionID
	Carrier     = wire.Carrier
	FlowHeader  = wire.FlowHeader
	Target      = wire.Target

	TCPConfig  = tcptls.Config
	TCPOptions = tcptls.TCPOptions
	TLSDialer  = tcptls.TLSDialer
)

var (
	NewCredentials = wire.NewCredentials
	NewTCPConfig   = tcptls.NewConfig
)

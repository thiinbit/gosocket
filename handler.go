// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import (
	"context"
	"net"
)

// ErrTimeout is returned for an expired deadline.
var ErrTimeout error = &TimeoutError{}
// TimeoutError is returned for an expired deadline.
type TimeoutError struct{}
// Implement the net.Error interface.
func (e *TimeoutError) Error() string   { return "i/o timeout" }

// PacketHandler on packet receive processor
type PacketHandler interface {
	PacketReceived(ctx context.Context, packet *Packet, session *Session)
	PacketSend(ctx context.Context, packet *Packet, session *Session)
}

// ClientPacketHandler
type ClientPacketHandler interface {
	PacketReceived(ctx context.Context, packet *Packet, cli *TCPClient)
	PacketSend(ctx context.Context, packet *Packet, cli *TCPClient)
}

// ConnectHandler on connect accept processor
type ConnectHandler interface {
	OnConnect(ctx context.Context, conn *net.TCPConn, tcpSer *TCPServer)
}

package gosocket

import (
	"context"
	"net"
)

// PacketHandler on packet receive processor
type PacketHandler interface {
	OnPacketReceived(ctx context.Context, packet *Packet, session *Session)
	OnPacketSend(ctx context.Context, packet *Packet, session *Session)
}

// ClientPacketHandler
type ClientPacketHandler interface {
	OnPacketReceived(ctx context.Context, packet *Packet, cli *TCPClient)
	OnPacketSend(ctx context.Context, packet *Packet, cli *TCPClient)
}

// ConnectHandler on connect accept processor
type ConnectHandler interface {
	OnConnect(ctx context.Context, conn *net.TCPConn, tcpSer *TCPServer)
}

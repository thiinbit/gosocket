package gosocket

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"io"
	"net"
	"time"
)

type defaultConnectHandler struct {
}

func (d defaultConnectHandler) OnConnect(ctx context.Context, conn *net.TCPConn, tcpSer *TCPServer) {
	s := NewSession(conn, tcpSer.defaultReadDeadline, tcpSer.defaultWriteDeadline, tcpSer.defaultHeartbeat, tcpSer)
	tcpSer.sessions[s.sID] = s
	tcpSer.debugLogger.Printf("Session create. sID: %s, client: %s", s.sID, s.conn.RemoteAddr().String())

	if tcpSer.sessionListener != nil {
		tcpSer.sessionListener.OnSessionCreate(s)
	}

	ctx2, cancel := context.WithCancel(ctx)

	go d.writeGo(ctx2, s, tcpSer)
	go d.readGo(ctx2, s, tcpSer)

	if <-s.closeSign {
		cancel()
		s.UpdateLastActive()

		delete(tcpSer.sessions, s.sID)

		if tcpSer.sessionListener != nil {
			tcpSer.sessionListener.OnSessionClose(s)
		}

		// Conn will close after return.
	}
}

func (d defaultConnectHandler) writeGo(ctx context.Context, s *Session, tcpSer *TCPServer) {
	for {
		select {

		// Cancel
		case <-ctx.Done():
			tcpSer.debugLogger.Printf("Session Write Done. sID: %s.", s.sID)
			return

		// Message write
		case msg := <-s.msgSendChan:

			data, err := tcpSer.codec.Encode(msg)
			if err != nil {
				s.CloseSession(fmt.Sprint("Encode data error.", err))
				return
			}

			size := uint32(len(data))
			if size > tcpSer.maxPacketBodyLen {
				s.CloseSession(fmt.Sprintf("Send packet size(%d) exceed max limit. ", size))
				return
			}

			pac := NewPacket(PacketVersion, size, data, adler32.Checksum(data))

			if err := s.conn.SetWriteDeadline(time.Now().Add(s.writeDeadline)); err != nil {
				s.CloseSession(fmt.Sprint("Set writeDeadline error.", err))
				return
			}

			tcpSer.packetHandler.OnPacketSend(ctx, pac, s)

		// Heartbeat
		case <-time.After(s.heartbeat):
			// Heartbeat can represent 256 instructions. 0: ping; 1: pong
			pingCmd := make([]byte, 1)
			pingCmd[0] = 0
			checksum := adler32.Checksum(pingCmd)

			pac := NewPacket(PacketHeartbeatVersion, 1, pingCmd, checksum)

			tcpSer.debugLogger.Printf("Heartbeat ping. sID: %s, checksum: %d", s.sID, checksum)
			tcpSer.packetHandler.OnPacketSend(ctx, pac, s)
		}
	}
}

func (d defaultConnectHandler) readGo(ctx context.Context, s *Session, tcpSer *TCPServer) {
	for {
		select {

		// Cancel
		case <-ctx.Done():
			tcpSer.debugLogger.Printf("Session Read Done. sID: %s.", s.sID)
			return

		// Message read
		default:
			if err := s.conn.SetReadDeadline(time.Now().Add(s.readDeadline)); err != nil {
				s.CloseSession(fmt.Sprint("Set ReadDeadline error.", err))
				return
			}

			// Read Version
			var verBuf [1]byte
			if _, err := s.conn.Read(verBuf[:]); err != nil {
				if err != io.EOF {
					s.CloseSession(fmt.Sprint("Read ver error.", err))
				} else {
					s.CloseSession(fmt.Sprint("Session EOF. "))
				}
				return
			}
			if verBuf[0] != PacketVersion && verBuf[0] != PacketHeartbeatVersion {
				s.CloseSession(fmt.Sprintf("Ver(%s) is wrong.", string(verBuf[0])))
			}

			// Read size
			var sizeBuf = make([]byte, 4)
			if i, err := s.conn.Read(sizeBuf); i < 4 || err != nil {
				s.CloseSession(fmt.Sprint("Read packet size error.", err))
				return
			}

			size := binary.BigEndian.Uint32(sizeBuf)
			if size > tcpSer.maxPacketBodyLen {
				s.CloseSession(fmt.Sprintf("Recv packet size(%d) exceed max limit. ", size))
				return
			}

			// Read body
			var dataBuf = make([]byte, size) // data size + checksum len
			if i, err := s.conn.Read(dataBuf); uint32(i) < size || err != nil {
				s.CloseSession(fmt.Sprint("Read packet body error.", err))
				return
			}

			// Read checksum
			var checksumBuf = make([]byte, 4)
			if i, err := s.conn.Read(checksumBuf); uint32(i) < 4 || err != nil {
				s.CloseSession(fmt.Sprint("Read packet checksum error. ", err))
				return
			}

			checksum := binary.BigEndian.Uint32(checksumBuf)
			packet := NewPacket(verBuf[0], size, dataBuf, checksum)

			if !packet.Checksum() {
				s.CloseSession(fmt.Sprint("Checksum error. Check false."))
				return
			}

			// Heartbeat or message receive
			if verBuf[0] == PacketHeartbeatVersion { // Heartbeat
				// Heartbeat can represent 256 instructions. 0: ping; 1: pong
				if len(dataBuf) == 1 && dataBuf[0] == 1 {
					tcpSer.debugLogger.Printf("Heartbeat pong. sID: %s, checksum: %d", s.sID, checksum)
				} else {
					tcpSer.debugLogger.Printf("Heartbeat unknown cmd. sID: %s, cmd: %s, checksum: %d", s.sID, string(dataBuf), checksum)
				}
			} else { // Message receive
				tcpSer.packetHandler.OnPacketReceived(ctx, packet, s)
			}
		}
	}
}

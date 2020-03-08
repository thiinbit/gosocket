// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

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

			tcpSer.packetHandler.PacketSend(ctx, pac, s)

		// Heartbeat
		case <-time.After(s.heartbeat):
			if s.lastActive.Add(s.heartbeat).After(time.Now()) {
				continue
			}

			// Heartbeat can represent 256 instructions. 0: ping; 1: pong
			pac := NewHeartbeatPacket(HeartbeatCmdPing)

			tcpSer.packetHandler.PacketSend(ctx, pac, s)
			tcpSer.debugLogger.Printf("Heartbeat ping sent. sID: %s, checksum: %d", s.sID, pac.checksum)
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
					s.CloseSession(fmt.Sprint("Read close. ", err))
				} else {
					s.CloseSession(fmt.Sprint("Session EOF. ", err))
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
				// Check heartbeat body length right and cmd in 0 or 1.
				if len(dataBuf) == 1 {
					if dataBuf[0] == HeartbeatCmdPong { // Received heartbeat pong
						tcpSer.debugLogger.Printf("Heartbeat pong received. sID: %s, checksum: %d", s.sID, checksum)
					}
					if dataBuf[0] == HeartbeatCmdPing { // Received heartbeat ping
						tcpSer.debugLogger.Printf("Heartbeat ping received. sID: %s, checksum: %d", s.sID, checksum)
						// Heartbeat can represent 256 instructions. 0: ping; 1: pong
						pac := NewHeartbeatPacket(HeartbeatCmdPong)

						tcpSer.packetHandler.PacketSend(ctx, pac, s)
						tcpSer.debugLogger.Printf("Heartbeat pong sent. sID: %s, checksum: %d", s.sID, pac.checksum)
					}
				} else {
					tcpSer.debugLogger.Printf("Heartbeat unknown cmd. sID: %s, cmd: %s, checksum: %d", s.sID, string(dataBuf), checksum)
				}
			} else { // Message receive
				tcpSer.packetHandler.PacketReceived(ctx, packet, s)
			}
		}
	}
}

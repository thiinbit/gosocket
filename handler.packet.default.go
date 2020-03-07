// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"time"
)

type defaultPacketHandler struct {
}

func (d defaultPacketHandler) OnPacketReceived(ctx context.Context, pac *Packet, s *Session) {

	// process chain if need extends

	m, err := s.serRef.codec.Decode(pac.body)

	if err != nil {
		s.CloseSession(fmt.Sprint("Packet decode error. ", err))
		return
	}

	s.serRef.debugLogger.Printf("Packet received: sID: %s, len: %d, checksum: %d", s.sID, pac.len, pac.checksum)

	s.UpdateLastActive()
	s.serRef.messageListener.OnMessage(ctx, m, s)
}

func (d defaultPacketHandler) OnPacketSend(ctx context.Context, pac *Packet, s *Session) {

	if err := s.conn.SetWriteDeadline(time.Now().Add(s.writeDeadline)); err != nil {
		s.CloseSession(fmt.Sprint("Set writeDeadline error.", err))
		return
	}

	// process chain if need extends

	dataBuf := new(bytes.Buffer)
	var errs [4]error

	errs[0] = binary.Write(dataBuf, binary.BigEndian, pac.ver)      // Ver 8 bit
	errs[1] = binary.Write(dataBuf, binary.BigEndian, pac.len)      // Size 32bit
	errs[2] = binary.Write(dataBuf, binary.BigEndian, pac.body)     // Data body len
	errs[3] = binary.Write(dataBuf, binary.BigEndian, pac.checksum) // Checksum 32bit

	for _, err := range errs {
		if err != nil {
			s.CloseSession(fmt.Sprintf("Packet to binary error. packetLen: %d. %v", pac.len, err))
			return
		}
	}

	s.serRef.debugLogger.Printf("Packet send: sID: %s, len: %d, checksum: %d", s.sID, pac.len, pac.checksum)

	s.UpdateLastActive()
	if i, err := s.conn.Write(dataBuf.Bytes()); err != nil {
		s.CloseSession(fmt.Sprintf("Packet write to socket error. writeLen: %d. %v", i, err))
		return
	}
}

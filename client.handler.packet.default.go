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

type defaultClientPacketHander struct {
}

func (d defaultClientPacketHander) PacketReceived(ctx context.Context, pac *Packet, cli *TCPClient) {

	// process chain if need extends

	m, err := cli.codec.Decode(pac.body)

	if err != nil {
		cli.Hangup(fmt.Sprint("Packet decode error.", err))
		return
	}

	cli.debugLogger.Printf("Client packet received. cli: %s, len: %d, checksum: %d.", cli.name, pac.len, pac.checksum)

	cli.UpdateLastActive()
	cli.messageListener.OnMessage(ctx, m, cli)
}

func (d defaultClientPacketHander) PacketSend(ctx context.Context, pac *Packet, cli *TCPClient) {

	// process chain if need extends

	if err := cli.connect.SetWriteDeadline(time.Now().Add(cli.writeDeadline)); err != nil {
		cli.Hangup(fmt.Sprint("setWriteDeadline error.", err))
		return
	}

	dataBuf := new(bytes.Buffer)
	var errs [4]error

	errs[0] = binary.Write(dataBuf, binary.BigEndian, pac.ver)      // Ver 8 bit
	errs[1] = binary.Write(dataBuf, binary.BigEndian, pac.len)      // Size 32bit
	errs[2] = binary.Write(dataBuf, binary.BigEndian, pac.body)     // Data body len
	errs[3] = binary.Write(dataBuf, binary.BigEndian, pac.checksum) // Checksum 32bit

	for _, err := range errs {
		if err != nil {
			cli.Hangup(fmt.Sprintf("Packet to binary error. packetLen: %d. %v", pac.len, err))
			return
		}
	}

	cli.debugLogger.Printf("Client packet send. cli: %s, len: %d, checksum: %d.", cli.name, pac.len, pac.checksum)

	if i, err := cli.connect.Write(dataBuf.Bytes()); err != nil {
		cli.Hangup(fmt.Sprintf("Packet write to socket error. writeLen: %d. %v", i, err))
		return
	}
	cli.UpdateLastActive()
}

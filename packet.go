// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import "hash/adler32"

const (
	PacketVersion          byte = 42
	PacketHeartbeatVersion byte = 255
)

const (
	// Default max packet body length 4MB
	defaultMaxPacketBodyLength = 4 * 1024 * 1024
)

type Packet struct {

	// ==== Header start ====
	// Version            //=
	ver byte ///////////////=
	// Packet length     //=
	len uint32 //////////////=
	// ==== Header end ======

	// ==== Body start ====
	// Packet body     //=
	body []byte //////////=
	// ==== Body end ======

	// ==== Checksum start ====
	// Checksum             //=
	checksum uint32 //////////=
	// ==== Checksum end ======
}

func NewPacket(ver byte, len uint32, packet []byte, checksum uint32) *Packet {
	return &Packet{
		ver:      ver,
		len:      len,
		body:     packet,
		checksum: checksum,
	}
}

// Ver return the packet version
func (p *Packet) Ver() byte {
	return p.ver
}

// Len return the packet body length
func (p *Packet) Len() uint32 {
	return p.len
}

// Body return the packet body
func (p *Packet) Body() []byte {
	return p.body
}

// Checksum return checksum is success
func (p *Packet) Checksum() bool {
	return p.checksum == adler32.Checksum(p.body)
}

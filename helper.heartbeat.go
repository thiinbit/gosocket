/*
 * // Copyright (c) 2020 @thiinbit. All rights reserved.
 * // Use of this source code is governed by an MIT-style
 * // license that can be found in the LICENSE file
 *
 */

package gosocket

import "hash/adler32"

// Build heartbeat packet. cmd -> 0: ping; 1: pong
// - see const HeartbeatCmdPing, HeartbeatCmdPong
func NewHeartbeatPacket(cmd byte) *Packet {
	// Heartbeat can represent 256 instructions. 0: ping; 1: pong
	cmdBody := make([]byte, 1)
	cmdBody[0] = cmd
	checksum := adler32.Checksum(cmdBody)

	return NewPacket(PacketHeartbeatVersion, 1, cmdBody, checksum)
}

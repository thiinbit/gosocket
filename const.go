/*
 * // Copyright (c) 2020 @thiinbit. All rights reserved.
 * // Use of this source code is governed by an MIT-style
 * // license that can be found in the LICENSE file
 *
 */

package gosocket

import "time"

// Server status
const (
	Preparing = "Preparing"
	Running   = "Running"
	Stop      = "Stop"
)

// Session status
const (
	statusCreated = "Created"
	statusClosed  = "Closed"
)

// Session keep alive const
const (
	sessionDefaultReadDeadline  = 5 * time.Second  // Default read deadline
	sessionDefaultWriteDeadline = 5 * time.Second  // Default Write deadline
	sessionDefaultHeartbeat     = 13 * time.Second // Default keepalive heart beat
)

// Send message channel const
const (
	defaultSendChanelCacheSize = 16
)

// Heartbeat cmd
const (
	HeartbeatCmdPing byte = 0
	HeartbeatCmdPong byte = 1
)

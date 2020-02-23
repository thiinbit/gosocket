// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import "context"

// MessageListener message processor interface
// Usage:
type MessageListener interface {
	OnMessage(ctx context.Context, message interface{}, session *Session)
}

type ClientMessageListener interface {
	OnMessage(ctx context.Context, message interface{}, cli *TCPClient)
}

type SessionListener interface {
	OnSessionCreate(session *Session)
	OnSessionClose(session *Session)
}

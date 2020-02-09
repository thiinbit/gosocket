package gosocket

import "context"

// MessageListener message processor interface
// Usage:
// *    TODO: write usage
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

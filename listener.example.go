package gosocket

import (
	"context"
	"log"
)

// ======== ======== Example server server message receive listener ======== ========
type ExampleServerMessageListener struct{}

func (e ExampleServerMessageListener) OnMessage(ctx context.Context, message interface{}, session *Session) {
	log.Printf("Received message from Client %s. Message content: %s.", session.RemoteAddr(), message)
}

// ======== ======== Example server session create/close listener ======== ========
type ExampleSessionListener struct{}

func (t ExampleSessionListener) OnSessionCreate(s *Session) {
	log.Printf("Server session create. sID: %s, remote: %s, createTime: %s, lastActive: %s",
		s.SID(), s.RemoteAddr(), s.CreateTime(), s.LastActive())
}

func (t ExampleSessionListener) OnSessionClose(s *Session) {
	log.Printf("Server session close. sID: %s, remote: %s, createTime: %s, lastActive: %s",
		s.SID(), s.RemoteAddr(), s.CreateTime(), s.LastActive())
}

// ======== ======== Example client message receive listener ======== ========
type ExampleClientMessageListener struct{}

func (t ExampleClientMessageListener) OnMessage(ctx context.Context, message interface{}, cli *TCPClient) {
	log.Printf("Received message from Server %s. Message content: %s.", cli.RemoteAddr(), message)
}

// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import (
	"context"
)

// ======== ======== Broadcast server message receive listener ======== ========
// BroadcastServerMessageListener
type BroadcastServerMessageListener struct{}

// BroadcastServerMessageListener impl
func (tl *BroadcastServerMessageListener) OnMessage(ctx context.Context, message interface{}, session *Session) {
	debugLog := session.serRef.debugLogger
	debugLog.Print("Server received message: ", Green(message))

	for k, v := range session.ServerRef().Sessions() {
		if session.SID() != k {
			debugLog.Printf("Broadcast message to client %s: %s ", Green(v.SID()), Green(message))
			v.SendMessage(message)
		}
	}
}

// ======== ======== Example server server message receive listener ======== ========
type ExampleServerMessageListener struct{}

func (e ExampleServerMessageListener) OnMessage(ctx context.Context, message interface{}, session *Session) {
	debugLog := session.serRef.debugLogger
	debugLog.Printf("Received message from Client %s. Message content: %s.", Green(session.RemoteAddr()), Green(message))
}

// ======== ======== Example server session create/close listener ======== ========
type ExampleSessionListener struct{}

func (t ExampleSessionListener) OnSessionCreate(s *Session) {
	debugLog := s.serRef.debugLogger
	debugLog.Printf("Server session create. sID: %s, remote: %s, createTime: %s, lastActive: %s",
		s.SID(), s.RemoteAddr(), s.CreateTime(), s.LastActive())
}

func (t ExampleSessionListener) OnSessionClose(s *Session) {
	debugLog := s.serRef.debugLogger
	debugLog.Printf("Server session close. sID: %s, remote: %s, createTime: %s, lastActive: %s",
		s.SID(), s.RemoteAddr(), s.CreateTime(), s.LastActive())
}

// ======== ======== Example client message receive listener ======== ========
type ExampleClientMessageListener struct{}

func (t ExampleClientMessageListener) OnMessage(ctx context.Context, message interface{}, cli *TCPClient) {
	cli.debugLogger.Printf("Received message from Server %s. Message content: %s.", Green(cli.RemoteAddr()), Green(message))
}

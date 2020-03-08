// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import (
	uuid "github.com/satori/go.uuid"
	"net"
	"sync"
	"time"
)

type SessionWriter interface {
	Write()
}

// ClientSession
type Session struct {
	sID           string
	status        string
	attributes    map[string]interface{}
	conn          *net.TCPConn
	readDeadline  time.Duration
	writeDeadline time.Duration
	heartbeat     time.Duration
	writer        *SessionWriter
	createTime    time.Time
	lastActive    time.Time
	serRef        *TCPServer
	closeSign     chan bool
	msgSendChan   chan interface{}
	mu            sync.Mutex
}

func NewSession(conn *net.TCPConn, readDeadline time.Duration, WriteDeadline time.Duration, heartbeat time.Duration, serverRef *TCPServer) *Session {
	return &Session{
		sID:           uuid.Must(uuid.NewV4()).String(),
		status:        statusCreated,
		attributes:    make(map[string]interface{}),
		conn:          conn,
		readDeadline:  readDeadline,
		writeDeadline: WriteDeadline,
		heartbeat:     heartbeat,
		createTime:    time.Now(),
		lastActive:    time.Now(),
		serRef:        serverRef,
		closeSign:     make(chan bool, 1),
		msgSendChan:   make(chan interface{}, defaultSendChanelCacheSize),
	}
}

func (s *Session) SendMessage(message interface{}) {
	s.msgSendChan <- message
}

func (s *Session) CloseSession(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != statusClosed {
		s.status = statusClosed
		s.closeSign <- true
		s.serRef.debugLogger.Printf(
			"Session close. sID: %s, cli: %s, reason: %s",
			s.sID, s.conn.RemoteAddr().String(), reason)
	}
}

// ServerRef return the server ref of session
func (s *Session) ServerRef() *TCPServer {
	return s.serRef
}

// SID return the session ID
func (s *Session) SID() string {
	return s.sID
}

// IsClosed return the session is closed
func (s *Session) IsClosed() bool {
	return s.status == statusClosed
}

// RemoteAddr return string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
func (s *Session) RemoteAddr() string {
	return s.conn.RemoteAddr().String()
}

// WriteDeadLine return the session write deadline
func (s *Session) WriteDeadline() time.Duration {
	return s.writeDeadline
}

func (s *Session) SetWriteDeadline(writeDeadline time.Duration) {
	s.writeDeadline = writeDeadline
}

func (s *Session) ReadDeadlin() time.Duration {
	return s.readDeadline
}
func (s *Session) SetReadDeadline(readDeadine time.Duration) {
	s.readDeadline = readDeadine
}

func (s *Session) Heartbeat() time.Duration {
	return s.readDeadline
}
func (s *Session) SetHeartbeat(heartbeat time.Duration) {
	s.heartbeat = heartbeat
}

// GetAttr get attribute by key
func (s *Session) Attr(key string) interface{} {
	return s.attributes[key]
}

// SetAttr set attribute key, value
func (s *Session) SetAttr(key string, val string) *Session {
	s.attributes[key] = val
	return s
}

// CreateTime return the session create time
func (s *Session) CreateTime() time.Time {
	return s.createTime
}

// LastActive return the session last active time. (update on create, close, send packet, receive packet)
func (s *Session) LastActive() time.Time {
	return s.lastActive
}

// UpdateLastActive update the session last active time.
func (s *Session) UpdateLastActive() {
	s.lastActive = time.Now()
}

// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPServer the tcp server struct
type TCPServer struct {
	env                  string              // Server Run environment DEBUG|RELEASE
	status               string              // Server status Preparing|Running|Stop
	sessions             map[string]*Session // Server connect sessions
	defaultReadDeadline  time.Duration       // Server session default read deadline (As default at session creation)
	defaultWriteDeadline time.Duration       // Server session default write deadline (As default at session creation)
	defaultHeartbeat     time.Duration       // Server session default heartbeat (As default at session creation)
	listener             *net.TCPListener    // Server net listener "127.0.0.1:5555" or "[::1]:8888"
	addr                 string              // Server listen address
	maxPacketBodyLen     uint32              // Server send/receive packet max body length limit (byte)
	debugLogger          DebugLogger         // Server debug logger
	logger               Logger              // Server run logger
	codec                Codec               // Server send/receive packet codec
	connectHandler       ConnectHandler      // Server new connect accept handler
	packetHandler        PacketHandler       // Server connect on packet receive handler
	messageListener      MessageListener     // Server message processor
	sessionListener      SessionListener     // Server session create/close listener
	stopSign             chan bool
	mu                   sync.Mutex
}

// NewTCPServer create a new tcp server
// Usage:
//  ============================================
//   ! Detail see: server_test.go OR README.md!
//  ============================================
//	// ==== ==== Server QuickStart ==== ====
//
//	// 1. Implement server OnMessageListener interface. see code above.
//	//     type MessageListener interface {
//	//         OnMessage(ctx context.Context, message interface{}, session *Session)
//	//     }
//
//	// 2. New TCPServer, register MessageListener to the server, and startup it.
//	//    - And now, congratulations! a tcp server is ready.
//	server, _ := NewTCPServer("[::1]:8888").
//		RegisterMessageListener(&TestExampleServerMessageListener{}). // Required
//		Run()
//
//	// 3. Stop the server when it is finished.
//	go func() {
//		<-time.NewTimer(10 * time.Second).C
//
//		server.Stop()
//	}()
//
//	// ==== ==== Client QuickStart ==== ====
//
//	// 1. Implement ClientMessageListener interface. see code above.
//	//     type ClientMessageListener interface {
//	//         OnMessage(ctx context.Context, message interface{}, cli *TCPClient)
//	//     }
//
//	// 2. New TCPClient, register ClientMessageListener to the client, and dial to server.
//	//    - And now, congratulations! a tcp client is ready.
//	client, _ := NewTcpClient("[::1]:8888").
//		RegisterMessageListener(&TestExampleClientListener{}).
//		Dial()
//
//	// 3. Say "Hello!" to server.
//	client.SendMessage("Hello!")
//
//	// 4. Hangup the client when it is finished.
//	go func() {
//		<-time.NewTimer(6 * time.Second).C
//
//		client.Hangup("It should be hangup now!")
//	}()
//  ============================================
//   ! Detail see: server_test.go OR README.md!
//  ============================================
func NewTCPServer(addr string) *TCPServer {
	return &TCPServer{
		env:                  DEBUG,
		status:               Preparing,
		sessions:             make(map[string]*Session),
		defaultWriteDeadline: sessionDefaultWriteDeadline,
		defaultReadDeadline:  sessionDefaultReadDeadline,
		defaultHeartbeat:     sessionDefaultHeartbeat,
		listener:             nil,
		addr:                 addr,
		maxPacketBodyLen:     defaultMaxPacketBodyLength,
		debugLogger:          DebugLogger{isDebugMode: true, logger: DefaultDebugLogger},
		logger:               DefaultLogger,
		codec:                DefaultCodec{},
		connectHandler:       defaultConnectHandler{},
		packetHandler:        defaultPacketHandler{},
		messageListener:      nil,
		sessionListener:      nil,
		stopSign:             make(chan bool),
	}
}

func (ts *TCPServer) Sessions() map[string]*Session {
	return ts.sessions
}

func (ts *TCPServer) RegisterMessageListener(listener MessageListener) *TCPServer {
	ts.checkPreparingStatus()
	ts.messageListener = listener
	return ts
}

func (ts *TCPServer) RegisterSessionListener(listener SessionListener) *TCPServer {
	ts.checkPreparingStatus()
	ts.sessionListener = listener
	return ts
}

func (ts *TCPServer) SetDebugMode(on bool) *TCPServer {
	ts.mu.Lock()

	if ts.env = DEBUG; !on {
		ts.env = RELEASE
	}
	ts.debugLogger.SetDebugMode(on)

	ts.mu.Unlock()

	return ts
}

func (ts *TCPServer) SetCodec(codec Codec) *TCPServer {
	ts.checkPreparingStatus()
	ts.codec = codec
	return ts
}

func (ts *TCPServer) SetMaxPacketBodyLength(maxLenBytes uint32) *TCPServer {
	ts.checkPreparingStatus()
	ts.maxPacketBodyLen = maxLenBytes
	return ts
}

func (ts *TCPServer) SetLogger(debugLogger Logger, logger Logger) *TCPServer {
	ts.checkPreparingStatus()
	ts.debugLogger = DebugLogger{isDebugMode: ts.debugLogger.isDebugMode, logger: debugLogger}
	ts.logger = logger
	return ts
}

//func (ts *TCPServer) SetConnectHandler(connHandler ConnectHandler) *TCPServer {
//	ts.checkPreparingStatus()
//	ts.connectHandler = connHandler
//	return ts
//}
//
//func (ts *TCPServer) SetPacketHandler(packetHander PacketHandler) *TCPServer {
//	ts.checkPreparingStatus()
//	ts.packetHandler = packetHander
//	return ts
//}

func (ts *TCPServer) SetHeartbeat(heartbeat time.Duration) *TCPServer {
	ts.checkPreparingStatus()
	ts.defaultHeartbeat = heartbeat
	ts.checkHeartbeatLtReadDead()
	return ts
}

// SetDefaultSessionReadDeadline session timeout if can not read any thing in this time
func (ts *TCPServer) SetDefaultSessionReadDeadline(read time.Duration) *TCPServer {
	ts.checkPreparingStatus()
	ts.defaultReadDeadline = read
	return ts
}

func (ts *TCPServer) SetDefaultSessionWriteDeadline(write time.Duration) *TCPServer {
	ts.checkPreparingStatus()
	ts.defaultWriteDeadline = write
	return ts
}

func (ts *TCPServer) Run() (*TCPServer, error) {
	// Message listener registered or panic
	ts.checkMessageListenerRegistered()

	var tcpAddr *net.TCPAddr
	var err error

	tcpAddr, err = net.ResolveTCPAddr("tcp", ts.addr)
	if err != nil {
		return nil, err
	}

	ts.listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	ts.mu.Lock()
	ts.status = Running
	ts.mu.Unlock()

	// Handle accept
	go ts.handleAccept(ctx)

	ts.logger.Printf("TCPServer run at %s.", ts.listener.Addr().String())

	// Stop holding
	go func() {
		<-ts.stopSign
		cancel()

		err := ts.listener.Close()
		if err != nil {
			ts.logger.Print("TCPServer close listen error. ", err)
		}

		ts.logger.Printf("TCPServer stop %s.", ts.listener.Addr().String())
	}()

	return ts, nil
}

func (ts *TCPServer) Stop() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.status != Stop {
		ts.status = Stop
		ts.stopSign <- true
	}

	return nil
}

func (ts *TCPServer) checkPreparingStatus() {
	if ts.status != Preparing {
		ts.logger.Panic("Can't change Server config on running or stop")
	}
}

func (ts *TCPServer) checkMessageListenerRegistered() {
	if ts.messageListener == nil {
		ts.logger.Panic("Message listener not registered!")
	}
}

func (ts *TCPServer) checkHeartbeatLtReadDead() {
	if ts.defaultHeartbeat >= ts.defaultReadDeadline {
		ts.logger.Panic("Heartbeat need time less than read deadline")
	}
}

func (ts *TCPServer) handleAccept(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			ts.logger.Println("Stop handle accept.")
			return

		default:
			conn, err := ts.listener.AcceptTCP()
			if err != nil {
				if fmt.Sprint(err.(*net.OpError).Err.Error()) == "use of closed network connection" {
					ts.debugLogger.Print("Accept closed")
				} else {
					ts.logger.Println("Handle accept failure: ", err)
				}
				continue
			}

			go func() {
				ts.connectHandler.OnConnect(ctx, conn, ts)

				if err := conn.Close(); err != nil {
					ts.logger.Printf("Conn close error. %v", err)
				} else {
					ts.debugLogger.Print("Conn close. ", conn.RemoteAddr().String())
				}
			}()
		}
	}
}

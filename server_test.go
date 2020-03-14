// Copyright 2020 @thiinbit. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file

package gosocket

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

// TestExampleServerMessageListener !required listener: listening server receives message.
type TestExampleServerMessageListener struct{}

func (tl *TestExampleServerMessageListener) OnMessage(ctx context.Context, message interface{}, session *Session) {
	log.Print("Server received message: ", message)

	// Reply to Client "Hi!" when received "Hello!".
	if message == "Hello!" || message == "Hello, Gosocket!" {
		session.SendMessage("Hi!")
	}
}

// TestExampleClientListener !required listener: listening client receives message.
type TestExampleClientListener struct{}

func (cl *TestExampleClientListener) OnMessage(ctx context.Context, message interface{}, cli *TCPClient) {
	log.Print("Client received message: ", message)

	// Reply to Server "Nice weather!" when received "Hi!"
	if message == "Hi!" {
		_ = cli.SendMessage("Nice weather!")
	}
}

// ======== ========             ======== ========
// ======== ======== Quick Start ======== ========
// ======== ========             ======== ========
func TestUsageQuickStart(t *testing.T) {
	//hold := make(chan bool, 1)

	// ==== ==== Server step ==== ====

	// 1. Implement server OnMessageListener interface. see code above.
	//     type MessageListener interface {
	//         OnMessage(ctx context.Context, message interface{}, session *Session)
	//     }

	// 2. New TCPServer, register MessageListener to the server, and startup it.
	//    - And now, congratulations! a tcp server is ready.
	server, _ := NewTCPServer("0.0.0.0:8888").
	//server, _ := NewTCPServer("[::1]:8888").
		RegisterMessageListener(&BroadcastServerMessageListener{}). // Required
		Run()

	//<-hold

	// 3. Stop the server when it is finished.
	go func() {
		<-time.NewTimer(20 * time.Second).C

		server.Stop()
	}()

	// ==== ==== Client step ==== ====

	// 1. Implement ClientMessageListener interface. see code above.
	//     type ClientMessageListener interface {
	//         OnMessage(ctx context.Context, message interface{}, cli *TCPClient)
	//     }

	// 2. New TCPClient, register ClientMessageListener to the client, and dial to server.
	//    - And now, congratulations! a tcp client is ready.
	client, _ := NewTcpClient("[::1]:8888").
		RegisterMessageListener(&TestExampleClientListener{}).
		Dial()

	// 3. Say "Hello!" to server.
	_ = client.SendMessage("Hello!")

	// 4. Hangup the client when it is finished.
	go func() {
		<-time.NewTimer(6 * time.Second).C

		client.Hangup("It should be hangup now!")
	}()

	// Hold test thread.
	<-time.NewTimer(13 * time.Second).C
}

// ======== ========                    ======== ========
// ======== ======== More Feature Below ======== ========
// ======== ========                    ======== ========

// TestExampleSessionListener !optional listener: listening server session create/close event.
// - When you want do something on session create/close.
type TestExampleSessionListener struct {
}

func (t TestExampleSessionListener) OnSessionCreate(s *Session) {
	log.Printf("Server session create. sID: %s, remote: %s, createTime: %s, lastActive: %s",
		s.SID(), s.RemoteAddr(), s.CreateTime(), s.LastActive())
}

func (t TestExampleSessionListener) OnSessionClose(s *Session) {
	log.Printf("Server session close. sID: %s, remote: %s, createTime: %s, lastActive: %s",
		s.SID(), s.RemoteAddr(), s.CreateTime(), s.LastActive())
}

type TestExampleCodec struct {
}

func (d TestExampleCodec) Encode(message interface{}) ([]byte, error) {
	// You can use JSON, protobuf and more other methods to serialize
	return []byte(fmt.Sprintf("%v", message)), nil
}

func (d TestExampleCodec) Decode(bytes []byte) (interface{}, error) {
	// You can use JSON, protobuf and more other methods to deserialize
	return string(bytes), nil
}

// ======== ========              ======== ========
// ======== ======== More Feature ======== ========
// ======== ========              ======== ========
func TestMoreFeatureUsage(t *testing.T) {
	// ==== ==== All features of the Server ==== ====
	server, _ := NewTCPServer("[::1]:8888").
		RegisterMessageListener(&TestExampleServerMessageListener{}). // Required: Listening receives message
		RegisterSessionListener(&TestExampleSessionListener{}). // Optional: Listening session create/close
		SetCodec(&TestExampleCodec{}). // Optional: Custom codec. Default codec directly to binary. You can choose to use JSON, protobuf and other methods you want to use.
		SetDebugMode(false). // Optional: The debug log will be printed in the DebugMode true, and the DebugMode false will not.
		// The parameters above need to be paid attention to, the parameters below do not need to be paid attention to.
		SetMaxPacketBodyLength(4*1024*1024). // Optional: Maximum bytes per message. Default 4M.
		SetHeartbeat(13*time.Second). // Optional: Heartbeat time. Default 13 seconds. Heartbeat only if no message is received. Heartbeat time must less than readDeadline!
		SetDefaultSessionReadDeadline(42*time.Second). // Optional: Read deadline time. Default 42 seconds. Time out automatically close session. It means that if the server don't receive any message or heartbeat for more than 42 seconds, will close the session.
		SetDefaultSessionWriteDeadline(5*time.Second). // Optional: Write deadline time. Default 5 seconds. If a message in the sending state is not sent for more than 5 seconds, the session will be automatically closed.
		//                                             // - In addition, the heartbeat/read/writeDeadline can be set individually for each session, and you can modify the heartbeat/readWriteDeadline of a single session at any time during runtime.
		SetLogger( // Optional: You can customize the logger. Compatible with go original log. Default is go original log with prefix [Gosocket]. You can use any log just implement these nine functions (Print(v ...interface{}), Printf(format string, v ...interface{}), Println(v ...interface{}), Fatal(v ...interface{}), Fatalf(format string, v ...interface{}), Fatalln(v ...interface{}), Panic(v ...interface{}), Panicf(format string, v ...interface{}), Panicln(v ...interface{})).
			log.New(os.Stderr, "[Gosocket-Debug]", log.LstdFlags), // Debug logger.
			log.New(os.Stderr, "[Gosocket]", log.LstdFlags)). // Release logger.
		Run() // Startup

	// Stop the server when it is finished.
	go func() {
		<-time.NewTimer(10 * time.Second).C

		server.Stop()
	}()

	// ==== ====  All features of the Client ==== ====
	client, _ := NewTcpClient("[::1]:8888").
		RegisterMessageListener(&TestExampleClientListener{}). // Required: Listening receives message
		SetCodec(&TestExampleCodec{}). // Optional: Custom codec. Default codec directly to binary. You can choose to use JSON, protobuf and other methods you want to use.
		// The parameters above need to be paid attention to, the parameters below do not need to be paid attention to.
		SetMaxPacketBodyLength(4*1024*1024).
		SetSessionReadDeadline(42*time.Second).
		SetSessionWriteDeadline(24*time.Hour).
		SetLogger( // Optional: You can customize the logger. Compatible with go original log. Default is go original log with prefix [Gosocket]. You can use any log just implement these nine functions (Print(v ...interface{}), Printf(format string, v ...interface{}), Println(v ...interface{}), Fatal(v ...interface{}), Fatalf(format string, v ...interface{}), Fatalln(v ...interface{}), Panic(v ...interface{}), Panicf(format string, v ...interface{}), Panicln(v ...interface{})).
			log.New(os.Stderr, "[Gosocket-Debug]", log.LstdFlags), // Debug logger.
			log.New(os.Stderr, "[Gosocket]", log.LstdFlags)). // Release logger.
		Dial()

	// 3. Say "Hello!" to server.
	_ = client.SendMessage("Hello!")

	// 4. Hangup the client when it is finished.
	go func() {
		<-time.NewTimer(6 * time.Second).C

		client.Hangup("It should be hangup!")
	}()

	// Hold test thread.
	<-time.NewTimer(13 * time.Second).C
}

// ==== ====                    ==== ====
// ==== ==== Test division line ==== ====
// ==== ====                    ==== ====

func TestServer(t *testing.T) {
	serverWait := 10 * time.Second
	clientWait := 6 * time.Second

	go newServer(serverWait, "[::1]:8888")
	go newClient(clientWait, "[::1]:8888")

	<-time.NewTimer(20 * time.Second).C
}

func newClient(runTime time.Duration, addr string) {
	client, err := NewTcpClient(addr).
		RegisterMessageListener(&TestExampleClientListener{}).
		Dial()

	if err != nil {
		log.Print(err)
	}

	log.Print("TestClient run.")
	_ = client.SendMessage("Hello!")

	<-time.NewTimer(runTime).C
	client.Hangup("TestClient should hangup.")
}

func newServer(runTime time.Duration, addr string) {
	serve, err := NewTCPServer(addr).
		SetHeartbeat(2 * time.Second). // Set heart beat 2 second. used to observe the heartbeat running state
		RegisterMessageListener(&TestExampleServerMessageListener{}). // Required
		RegisterSessionListener(&TestExampleSessionListener{}). // Optional
		Run()

	if err != nil {
		log.Print(err)
	}

	log.Print("TestServer run. sessions: ", len(serve.Sessions()))

	<-time.NewTimer(runTime).C
	if err := serve.Stop(); err != nil {
		log.Print(err)
	}
}

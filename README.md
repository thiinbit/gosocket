# Gosocket


> Gosocket is a simple, lightweight, session, heartbeat socket library written in Go (Golang). Supports TCP now. UDP and WS will be supported in future. If you need small and simple enough, you will love Gosocket.

## Example client
![demo](https://github.com/thiinbit/gosocket/blob/master/cli/demo-1920x730.gif)

Build example client
```sh
cd gosocket/cli
./buildDarwin.sh
``` 
Start server
```sh
cd gosocket/cli
./gosocket_darwin_amd64 server -l 0.0.0.0:8888 -d true
``` 
Start client
```sh
cd gosocket/cli
./gosocket_darwin_amd64 client -t 127.0.0.1:8888 -d true
``` 

## Installation

To install Gin package, you need to install Go and set your Go workspace first.

1. The first need [Go](https://golang.org/) installed (**version 1.11+ is required**), then you can use the below Go command to install Gin.

```sh
$ go get -u github.com/thiinbit/gosocket
```

2. Import it in your code:

```go
import "github.com/thiinbit/gosocket"
```

## QuickStart 
(see server_test.go -> func TestUsageQuickStart)

### Part.1 Create Server step.
1. Create server message listener. (Process on message received from client.)

```go
// Implement server OnMessageListener interface. Just like code below. 
//     type MessageListener interface {
//         OnMessage(ctx context.Context, message interface{}, session *Session)
//     }

// TestExampleServerMessageListener !required listener: listening server receives message.
type TestExampleServerMessageListener struct{}

func (tl *TestExampleServerMessageListener) OnMessage(ctx context.Context, message interface{}, session *Session) {
	log.Print("Server received message: ", message)

	// Reply to Client "Hi!" when received "Hello!".
	if message == "Hello!" {
		session.SendMessage("Hi!")
	}
}
```

2. Create TCPServer and run it.
```go
	// New TCPServer, register MessageListener to the server, and startup it.
	server, _ := NewTCPServer("[::1]:8888").
		RegisterMessageListener(&TestExampleServerMessageListener{}). // Required
		Run()

	//   - And now, congratulations! a tcp server is ready.

	//  Stop the server when it is finished.
	go func() {
		<-time.NewTimer(10 * time.Second).C

		server.Stop()
	}()
``` 

### Part.2 Create Client

1. Create client message listener. (Process on message received from server.)
```go
// Implement ClientMessageListener interface. Just like code below.
//     type ClientMessageListener interface {
//         OnMessage(ctx context.Context, message interface{}, cli *TCPClient)
//     }

// TestExampleClientListener !required listener: listening client receives message.
type TestExampleClientListener struct{}

func (cl *TestExampleClientListener) OnMessage(ctx context.Context, message interface{}, cli *TCPClient) {
	log.Print("Client received message: ", message)

	// Reply to Server "Nice weather!" when received "Hi!"
	if message == "Hi!" {
		cli.SendMessage("Nice weather!")
	}
}

2. Create TCPClient, dial to server, and say "Hello!".
	// New TCPClient, register ClientMessageListener to the client, and dial to server.
	client, _ := NewTcpClient("[::1]:8888").
		RegisterMessageListener(&TestExampleClientListener{}).
		Dial()

    //   - And now, congratulations! a tcp client is ready.

	// Say "Hello!" to server.
	client.SendMessage("Hello!")

	// Hangup the client when it is finished.
	go func() {
		<-time.NewTimer(10 * time.Second).C

		client.Hangup("It should be hangup now!")
	}()
```


## More Usage
// TODO: Supplement more detailed documents

### All features of the Server
```go
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
```

### All features of the Client
```go
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
```


### Custom Code and Session Listener example.
```go
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

// TestExampleCodec you can custom your owner message codec, like use JSON, protobuf, more and more.
// - a plain string codec.
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
``` 

## Client SDKs

- Swift client SDK: [Gosocket-Swift](https://github.com/thiinbit/Gosocket-Swift) 
- Java client SDK: [gosocket4j](https://github.com/thiinbit/gosocket4j) 

### TODO:
1. Supplement more detailed documents and use cases
2. Swift client sdk -> first version done.
3. Java client sdk -> first version done.
4. Support UDP, WEBSOCKET.
5. Keep it's simple.


### Ver:
0.0.1:
- First version. Base functions.  TCP message, heartbeat, session management, and more.
0.0.2:
- Add a client example.
- Some fix, adj.
    
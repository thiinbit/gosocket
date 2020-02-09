package gosocket

import (
	"context"
	"encoding/binary"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"hash/adler32"
	"io"
	"net"
	"sync"
	"time"
)

// TCPClient the tcp server struct
type TCPClient struct {
	name             string              // Client Name
	env              string              // Client Run environment DEBUG|RELEASE
	status           string              // Client status Preparing|Running|Stop
	readDeadline     time.Duration       // Client read deadline
	writeDeadline    time.Duration       // Client write deadline
	connect          *net.TCPConn        // Client TCP conn
	serverAddr       string              // Client connect server address ("tcp", "golang.org:http"| "tcp", "198.51.100.1:80 | [fe80::1%lo0]:53)
	maxPacketBodyLen uint32              // Client send/receive packet max body length limit (byte)
	debugLogger      DebugLogger         // Client debug logger
	logger           Logger              // Client run logger
	codec            Codec               // Client send/receive packet codec
	packetHandler    ClientPacketHandler // Client connect on packet receive handler
	messageListener  ClientMessageListener
	hangupSign       chan bool
	msgSendChan      chan interface{}
	mu               sync.Mutex
}

// NewTcpClient create a new tcp server
// Usage:
// *    TODO: write usage
func NewTcpClient(serAddr string) *TCPClient {
	return &TCPClient{
		name:             uuid.Must(uuid.NewV4()).String(),
		env:              DEBUG,
		status:           Preparing,
		writeDeadline:    24 * time.Hour, // Client write deadline is 24 hour(need reconnect after one day)
		readDeadline:     sessionDefaultReadDeadline,
		connect:          nil,
		serverAddr:       serAddr,
		maxPacketBodyLen: defaultMaxPacketBodyLength,
		debugLogger:      DebugLogger{isDebugMode: true, logger: DefaultDebugLogger},
		logger:           DefaultLogger,
		codec:            DefaultCodec{},
		packetHandler:    defaultClientPacketHander{},
		messageListener:  nil,
		hangupSign:       make(chan bool),
		msgSendChan:      make(chan interface{}, 8),
	}
}

func (cli *TCPClient) RegisterMessageListener(listener ClientMessageListener) *TCPClient {
	cli.checkPreparingStatus()
	cli.messageListener = listener
	return cli
}

func (cli *TCPClient) SetMaxPacketBodyLength(maxMsgBodyLen uint32) *TCPClient {
	cli.checkPreparingStatus()
	cli.maxPacketBodyLen = maxMsgBodyLen
	return cli
}

func (cli *TCPClient) SetLogger(debugLogger Logger, logger Logger) *TCPClient {
	cli.mu.Lock()
	cli.debugLogger = DebugLogger{isDebugMode: cli.debugLogger.isDebugMode, logger: debugLogger}
	cli.logger = logger
	cli.mu.Unlock()
	return cli
}

func (cli *TCPClient) SetCodec(codec Codec) *TCPClient {
	cli.checkPreparingStatus()
	cli.codec = codec
	return cli
}

//func (cli *TCPClient) SetPacketHandler(packetHandler ClientPacketHandler) *TCPClient {
//	cli.checkPreparingStatus()
//	cli.packetHandler = packetHandler
//	return cli
//}

func (cli *TCPClient) SetSessionReadDeadline(read time.Duration) *TCPClient {
	cli.checkPreparingStatus()
	cli.readDeadline = read
	return cli
}

func (cli *TCPClient) SetSessionWriteDeadline(write time.Duration) *TCPClient {
	cli.checkPreparingStatus()
	cli.writeDeadline = write
	return cli
}

func (cli *TCPClient) RemoteAddr() string {
	return cli.connect.RemoteAddr().String()
}

func (cli *TCPClient) checkPreparingStatus() {
	if cli.status != Preparing {
		cli.logger.Panic("Can't change Client config on running or stop")
	}
}

func (cli *TCPClient) checkMessageListenerRegistered() {
	if cli.messageListener == nil {
		cli.logger.Panic("Message listener not registered!")
	}
}

func (cli *TCPClient) Dial() (*TCPClient, error) {
	// Message listener registered or panic
	//cli.checkMessageListenerRegistered()

	var tcpAddr *net.TCPAddr
	var err error

	tcpAddr, err = net.ResolveTCPAddr("tcp", cli.serverAddr)
	if err != nil {
		return nil, err
	}

	cli.connect, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	cli.mu.Lock()
	cli.status = Running
	cli.mu.Unlock()

	// Handle connect
	go cli.handleConnect(ctx)

	cli.logger.Printf("TCPClient dialed %s.", cli.connect.RemoteAddr().String())

	// Stop holding
	go func() {
		<-cli.hangupSign
		cancel()

		err := cli.connect.Close()
		if err != nil {
			cli.logger.Print("Close connect error.", err)
		}

		cli.logger.Printf("TCPClient %s hangup %s.", cli.name, cli.connect.RemoteAddr().String())
	}()

	return cli, nil
}

func (cli *TCPClient) SendMessage(msg interface{}) {
	cli.msgSendChan <- msg
}

func (cli *TCPClient) Hangup(reason string) {
	cli.mu.Lock()
	defer cli.mu.Unlock()

	if cli.status != Stop {
		cli.status = Stop
		cli.hangupSign <- true
		cli.debugLogger.Printf("Client hangup %s on %s->%s. reason: %s",
			cli.name, cli.connect.LocalAddr().String(), cli.connect.RemoteAddr().String(), reason)
	}
}

func (cli *TCPClient) handleConnect(ctx context.Context) {
	go cli.handleWrite(ctx)
	go cli.handleRead(ctx)
}

func (cli *TCPClient) handleWrite(ctx context.Context) {
	for {
		select {

		case <-ctx.Done():
			cli.logger.Println("Client stop handle write.")
			return

		case msg := <-cli.msgSendChan:

			data, err := cli.codec.Encode(msg)

			if err != nil {
				cli.Hangup(fmt.Sprint("encode data error.", err))
				return
			}

			size := uint32(len(data))
			if size > cli.maxPacketBodyLen {
				cli.Hangup(fmt.Sprintf("Send packet size(%d) exceed max limit. ", size))
				return
			}

			pac := NewPacket(PacketVersion, size, data, adler32.Checksum(data))

			cli.packetHandler.OnPacketSend(ctx, pac, cli)

		case <-time.After(5 * time.Second):
			cli.debugLogger.Printf("Cli %s healthy check.", cli.name)
		}
	}
}

func (cli *TCPClient) handleRead(ctx context.Context) {
	for {
		select {

		case <-ctx.Done():
			cli.logger.Println("Client stop handle read.")
			return

		default:
			if err := cli.connect.SetReadDeadline(time.Now().Add(cli.readDeadline)); err != nil {
				cli.Hangup(fmt.Sprint("Set ReadDeadline error. ", err))
				return
			}

			// Read Version
			var verBuf [1]byte
			if _, err := cli.connect.Read(verBuf[:]); err != nil {
				if err != io.EOF {
					cli.Hangup(fmt.Sprint("Read ver error. ", err))
				} else {
					cli.Hangup(fmt.Sprint("EOF. ", err))
				}
				return
			}
			if verBuf[0] != PacketVersion && verBuf[0] != PacketHeartbeatVersion {
				cli.Hangup(fmt.Sprintf("Ver(%s) is wrong.", string(verBuf[0])))
				return
			}

			// Read size
			var sizeBuf = make([]byte, 4)
			if i, err := cli.connect.Read(sizeBuf); i < 4 || err != nil {
				cli.Hangup(fmt.Sprint("Read packet size error. ", err))
				return
			}

			size := binary.BigEndian.Uint32(sizeBuf)
			if size > cli.maxPacketBodyLen {
				cli.Hangup(fmt.Sprintf("Recv packet size(%d) exceed max limit. ", size))
				return
			}

			var dataBuf = make([]byte, size) // data size + checksum len
			if i, err := cli.connect.Read(dataBuf); uint32(i) < size || err != nil {
				cli.Hangup(fmt.Sprint("Read packet body err. ", err))
				return
			}

			var checksumBuf = make([]byte, 4)
			if i, err := cli.connect.Read(checksumBuf); uint32(i) < 4 || err != nil {
				cli.Hangup(fmt.Sprint("Read packet checksum err. ", err))
				return
			}

			checksum := binary.BigEndian.Uint32(checksumBuf)
			packet := NewPacket(verBuf[0], size, dataBuf, checksum)

			if !packet.Checksum() {
				cli.Hangup(fmt.Sprint("Checksum err. Check false."))
				return
			}

			// Heartbeat or message
			if verBuf[0] == PacketHeartbeatVersion { // Heartbeat
				// Heartbeat can represent 256 instructions. 0: ping; 1: pong
				pongCmd := make([]byte, 1)
				pongCmd [0] = 1
				checksum := adler32.Checksum(pongCmd)

				pac := NewPacket(PacketHeartbeatVersion, 1, pongCmd, checksum)

				cli.debugLogger.Printf("Heartbeat pong. cli: %s, checksum: %d", cli.name, checksum)
				cli.packetHandler.OnPacketSend(ctx, pac, cli)
			} else { // Message
				cli.packetHandler.OnPacketReceived(ctx, packet, cli)
			}
		}
	}
}

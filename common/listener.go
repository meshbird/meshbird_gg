package common

import (
	"fmt"
	"github.com/gophergala2016/meshbird/network/protocol"
	"log"
	"net"
	"os"
)

type ListenerService struct {
	BaseService

	localNode *LocalNode
	listener  net.Listener

	logger *log.Logger
}

func (l ListenerService) Name() string {
	return "listener"
}

func (l *ListenerService) Init(ln *LocalNode) error {
	l.logger = log.New(os.Stderr, "[listener] ", log.LstdFlags)
	l.logger.Printf("Listening on port: %d", ln.State().ListenPort+1)

	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", ln.State().ListenPort+1))
	if err != nil {
		return err
	}

	l.localNode = ln
	l.listener = listener
	return nil
}

func (l *ListenerService) Run() error {
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			break
		}

		l.logger.Printf("Has new connection: %s", conn.RemoteAddr().String())

		if err = l.process(conn); err != nil {
			l.logger.Printf("Error on process: %s", err)
		}
	}
	return nil
}

func (l *ListenerService) Stop() {
	l.SetStatus(StatusStopping)
	l.listener.Close()
}

func (l *ListenerService) process(c net.Conn) error {
	//defer c.Close()

	handshakeMsg, errHandshake := protocol.ReadDecodeHandshake(c)
	if errHandshake != nil {
		return errHandshake
	}

	l.logger.Println("Processing hansdhake...")

	if !protocol.IsMagicValid(handshakeMsg.Bytes()) {
		return fmt.Errorf("Invalid magic bytes")
	}

	l.logger.Println("Magic bytes are correct. Preparing reply...")

	if err := protocol.WriteEncodeOk(c); err != nil {
		return err
	}
	if err := protocol.WriteEncodePeerInfo(c, l.localNode.State().PrivateIP); err != nil {
		return err
	}

	peerInfo, errPeerInfo := protocol.ReadDecodePeerInfo(c)
	if errPeerInfo != nil {
		return errPeerInfo
	}

	l.logger.Println("Processing PeerInfo...")

	rn := NewRemoteNode(c, handshakeMsg.SessionKey(), peerInfo.PrivateIP())

	netTable, ok := l.localNode.Service("net-table").(*NetTable)
	if !ok || netTable == nil {
		return fmt.Errorf("net-table is nil")
	}

	l.logger.Printf("Adding remote node from listener...")
	netTable.AddRemoteNode(rn)

	return nil
}

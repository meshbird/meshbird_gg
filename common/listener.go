package common

import (
	"fmt"
	//"github.com/anacrolix/utp"
	"bytes"
	"github.com/gophergala2016/meshbird/network/protocol"
	"log"
	"net"
)

type ListenerService struct {
	BaseService

	localNode *LocalNode
	//socket    *utp.Socket
	conn *net.UDPConn
}

func (l ListenerService) Name() string {
	return "listener"
}

func (l *ListenerService) Init(ln *LocalNode) error {
	log.Printf("Listening on port: %d", ln.State().ListenPort+1)
	//socket, err := utp.NewSocket("udp4", fmt.Sprintf("0.0.0.0:%d", ln.State().ListenPort+1))

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: ln.State().ListenPort + 1})
	if err != nil {
		return err
	}

	l.localNode = ln
	//l.socket = socket
	l.conn = conn
	return nil
}

func (l *ListenerService) Run() error {
	for {
		buf := make([]byte, 1500)
		n, addr, err := l.conn.ReadFrom(buf)
		//conn, err := l.socket.Accept()
		if err != nil {
			log.Printf("Error on read from UDP: %v", err)
			break
		}

		log.Printf("Has new connection: %s", addr.String())

		if err = l.process(buf[:n], addr); err != nil {
			log.Printf("Error on process: %s", err)
		}
	}
	return nil
}

func (l *ListenerService) Stop() {
	l.SetStatus(StatusStopping)
	//l.socket.Close()
	l.conn.Close()
}

func (l *ListenerService) Conn() *net.UDPConn {
	return l.conn
}

func (l *ListenerService) process(payload []byte, addr net.Addr) error {
	//defer c.Close()
	c := bytes.NewBuffer(payload)

	handshakeMsg, errHandshake := protocol.ReadDecodeHandshake(c, nil)
	if errHandshake != nil {
		return errHandshake
	}

	log.Println("Processing hansdhake...")

	if !protocol.IsMagicValid(handshakeMsg.Bytes()) {
		return fmt.Errorf("Invalid magic bytes")
	}

	log.Println("Magic bytes are correct. Preparing reply...")

	if err := protocol.WriteEncodeOk(c); err != nil {
		return err
	}
	if err := protocol.WriteEncodePeerInfo(c, l.localNode.State().PrivateIP); err != nil {
		return err
	}

	peerInfo, errPeerInfo := protocol.ReadDecodePeerInfo(c, nil)
	if errPeerInfo != nil {
		return errPeerInfo
	}

	log.Println("Processing PeerInfo...")

	rn := NewRemoteNode(handshakeMsg.SessionKey(), peerInfo.PrivateIP(), addr)

	netTable, ok := l.localNode.Service("net-table").(*NetTable)
	if !ok || netTable == nil {
		return fmt.Errorf("net-table is nil")
	}

	netTable.AddRemoteNode(rn)

	return nil
}

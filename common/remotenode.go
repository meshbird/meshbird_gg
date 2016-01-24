package common

import (
	"fmt"
	//"github.com/anacrolix/utp"
	"log"
	"net"
	"strconv"
	//"time"

	"github.com/gophergala2016/meshbird/network/protocol"
	"github.com/gophergala2016/meshbird/secure"
)

type RemoteNode struct {
	Node
	sessionKey    []byte
	privateIP     net.IP
	publicAddress *net.UDPAddr
}

func NewRemoteNode(sessionKey []byte, privateIP net.IP, publicAddress *net.UDPAddr) *RemoteNode {
	// TODO: Conn
	return &RemoteNode{
		sessionKey:    sessionKey,
		privateIP:     privateIP,
		publicAddress: publicAddress,
	}
}

func (rn *RemoteNode) SendPacket(dstIP net.IP, payload []byte) error {
	return nil
}

func TryConnect(ln *LocalNode, h string, networkSecret *secure.NetworkSecret) (*RemoteNode, error) {
	host, portStr, errSplit := net.SplitHostPort(h)
	if errSplit != nil {
		return nil, errSplit
	}

	port, errConvert := strconv.Atoi(portStr)
	if errConvert != nil {
		return nil, errConvert
	}

	rn := new(RemoteNode)
	rn.publicAddress = &net.UDPAddr{IP: host, Port: port + 1}

	rn.sessionKey = RandomBytes(16)

	if err := protocol.WriteEncodeHandshake(rn.conn, rn.sessionKey, networkSecret); err != nil {
		return nil, err
	}
	if _, okError := protocol.ReadDecodeOk(rn.conn, rn.sessionKey); okError != nil {
		return nil, okError
	}

	peerInfo, errPeerInfo := protocol.ReadDecodePeerInfo(rn.conn, rn.sessionKey)
	if errPeerInfo != nil {
		return nil, errPeerInfo
	}

	rn.privateIP = peerInfo.PrivateIP()

	if err := protocol.WriteEncodePeerInfo(rn.conn, rn.privateIP); err != nil {
		return nil, err
	}

	log.Printf("Connected to node: %s/%s", rn.privateIP.String(), rn.publicAddress.String())

	return rn, nil
}

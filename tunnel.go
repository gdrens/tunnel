package tunnel

import (
	"crypto/md5"
	"fmt"
	"log"
	"net"
)

type TunConn struct {
	net.Conn
}

var _ net.Conn = (*TunConn)(nil)

func CreateTunnel(conn net.Conn, key string) (*TunConn, error) {
	chaConn, err := NewCha20Conn(conn, key)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	cmpConn := NewCmpConn(chaConn)
	return &TunConn{
		Conn: cmpConn,
	}, nil
}

func generateKey(key string) []byte {
	//generate 32 bytes key
	return []byte(fmt.Sprintf("%x", md5.Sum([]byte(key))))
}

package tunnel

import (
	"crypto/rand"
	"errors"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/chacha20"
)

type Cha20Conn struct {
	key     []byte //Must be 32 bytes
	encoder *chacha20.Cipher
	decoder *chacha20.Cipher
	net.Conn
}

var _ net.Conn = (*Cha20Conn)(nil)

func NewCha20Conn(conn net.Conn, key string) (*Cha20Conn, error) {
	k := generateKey(key)
	c := &Cha20Conn{
		key:     k,
		Conn:    conn,
		encoder: encoder(conn, k),
		decoder: decoder(conn, k),
	}
	if c.decoder == nil || c.encoder == nil {
		return nil, errors.New("create chacha20 stream cipher false")
	}
	return c, nil
}

func (c *Cha20Conn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n == 0 {
		return n, err
	}
	dst := make([]byte, n)
	pn := p[:n]
	c.decoder.XORKeyStream(dst, pn)
	copy(pn, dst)
	return n, err
}

func (c *Cha20Conn) Write(p []byte) (int, error) {
	dst := make([]byte, len(p))
	c.encoder.XORKeyStream(dst, p)
	return c.Conn.Write(dst)
}

func decoder(con net.Conn, key []byte) *chacha20.Cipher {
	nonce := make([]byte, chacha20.NonceSizeX)
	if _, err := io.ReadFull(con, nonce); err != nil {
		log.Print(err)
		return nil
	}
	decoder, _ := chacha20.NewUnauthenticatedCipher(key, nonce)
	return decoder
}

func encoder(con net.Conn, key []byte) *chacha20.Cipher {
	nonce := make([]byte, chacha20.NonceSizeX)
	rand.Read(nonce)
	if _, err := con.Write(nonce); err != nil {
		log.Print(err)
		return nil
	}
	cipher, _ := chacha20.NewUnauthenticatedCipher(key, nonce)
	return cipher
}

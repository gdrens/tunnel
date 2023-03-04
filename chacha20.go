package tunnel

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20"
)

type Chacha20 struct {
	encoder *chacha20.Cipher
	decoder *chacha20.Cipher
	key     []byte
	conn    io.ReadWriteCloser
}

func NewChacha20(conn io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	chacha20 := &Chacha20{
		key:  generateKey32(cfg.Key),
		conn: conn,
	}
	if err := chacha20.createEncoder(); err != nil {
		return nil, err
	}
	if err := chacha20.createDecoder(); err != nil {
		return nil, err
	}
	return chacha20, nil
}

func (c *Chacha20) Read(p []byte) (int, error) {
	n, err := c.conn.Read(p)
	if err != nil {
		return n, err
	}
	dst := make([]byte, n)
	pn := p[:n]
	c.decoder.XORKeyStream(dst, pn)
	copy(pn, dst)
	return n, err
}

func (c *Chacha20) Write(p []byte) (int, error) {
	dst := make([]byte, len(p))
	c.encoder.XORKeyStream(dst, p)
	return c.conn.Write(dst)
}

func (c *Chacha20) Close() error {
	return c.conn.Close()
}

func (c *Chacha20) createDecoder() error {
	nonce := make([]byte, chacha20.NonceSizeX)
	n, err := c.conn.Read(nonce)
	if err != nil {
		return err
	}
	if n != chacha20.NonceSizeX {
		return errors.New("Could not read nonce from the connection")
	}
	cipher, err := chacha20.NewUnauthenticatedCipher(c.key, nonce)
	if err != nil {
		return err
	}
	c.decoder = cipher
	return nil
}

func (c *Chacha20) createEncoder() error {
	nonce := make([]byte, chacha20.NonceSizeX)
	rand.Read(nonce)
	if _, err := c.conn.Write(nonce); err != nil {
		return err
	}
	cipher, err := chacha20.NewUnauthenticatedCipher(c.key, nonce)
	if err != nil {
		return err
	}
	c.encoder = cipher
	return nil
}

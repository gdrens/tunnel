package tunnel

import (
	"compress/flate"
	"io"
	"log"
	"net"
)

type CompressConn struct {
	r io.Reader
	w *flate.Writer
	net.Conn
}

var _ net.Conn = (*CompressConn)(nil)

func NewCmpConn(conn net.Conn) (*CompressConn, error) {
	r := flate.NewReader(conn)
	w, err := flate.NewWriter(conn, flate.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &CompressConn{
		r:    r,
		w:    w,
		Conn: conn,
	}, nil
}

func (c *CompressConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

func (c *CompressConn) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if err := c.w.Flush(); err != nil {
		log.Print(err)
		return 0, err
	}
	return n, err
}

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

func NewCmpConn(conn net.Conn) *CompressConn {
	r := flate.NewReader(conn)
	w, _ := flate.NewWriter(conn, flate.BestSpeed)
	return &CompressConn{
		r:    r,
		w:    w,
		Conn: conn,
	}
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

package tunnel

import (
	"compress/flate"
	"io"
	"net"
	"time"
)

type Compression struct {
	r    io.ReadCloser
	w    *flate.Writer
	conn net.Conn
}

var _ net.Conn = (*Compression)(nil)

func NewCompress(conn net.Conn, level int) (net.Conn, error) {
	read := flate.NewReader(conn)
	write, err := flate.NewWriter(conn, level)
	if err != nil {
		return nil, err
	}
	compress := &Compression{
		r:    read,
		w:    write,
		conn: conn,
	}
	return compress, nil
}

func (c *Compression) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

func (c *Compression) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if err != nil {
		return 0, err
	}
	c.w.Flush()
	return n, err
}

func (c *Compression) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Compression) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *Compression) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Compression) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Compression) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Compression) Close() error {
	c.r.Close()
	c.w.Close()
	return c.conn.Close()
}

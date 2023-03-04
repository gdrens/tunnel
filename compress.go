package tunnel

import (
	"compress/flate"
	"io"
)

type Compression struct {
	r    io.ReadCloser
	w    *flate.Writer
	conn io.ReadWriteCloser
}

func NewCompress(conn io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	read := flate.NewReader(conn)
	write, err := flate.NewWriter(conn, CompressLevel)
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

func (c *Compression) Close() error {
	c.r.Close()
	c.w.Close()
	return c.conn.Close()
}

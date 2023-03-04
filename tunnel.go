package tunnel

import (
	"io"
	"log"
	"net"
	"time"
)

type Tunnel struct {
	listen net.Listener
	hook   func(dst, src io.ReadWriteCloser)
}

const CompressLevel = 9

var cfg *Config

//  NoCompression      = 0
//	BestSpeed          = 1
//	BestCompression    = 9
//	DefaultCompression = -1

type Stream func(io.ReadWriteCloser) (io.ReadWriteCloser, error)

type Config struct {
	NetWork        string
	DialAddrPort   string
	ListenAddrPort string
	IsClient       bool
	Key            string
}

var defaultUints = []Stream{NewCompress, NewChacha20}

func New(conf *Config) (*Tunnel, error) {
	cfg = conf
	if cfg.NetWork == "" {
		cfg.NetWork = "tcp"
	}
	l, err := net.Listen(cfg.NetWork, cfg.ListenAddrPort)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &Tunnel{
		listen: l,
	}, nil
}

func (t *Tunnel) Run() {
	if cfg.IsClient {
		client(t)
	} else {
		server(t)
	}
}

func (t *Tunnel) Register(hook func(dst, src io.ReadWriteCloser)) {
	t.hook = hook
}

func (t *Tunnel) New() (io.ReadWriteCloser, error) {
	c, err := net.DialTimeout(cfg.NetWork, cfg.DialAddrPort, 15*time.Second)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	conn, err := createStream(c)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func createStream(conn io.ReadWriteCloser) (tConn io.ReadWriteCloser, err error) {
	tConn = conn
	for i := 0; i < len(defaultUints); i++ {
		tConn, err = defaultUints[i](tConn)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

func Forward(dst, src io.ReadWriteCloser) {
	defer dst.Close()
	defer src.Close()
	go func() {
		io.Copy(dst, src)
		dst.Close()
		src.Close()
	}()
	io.Copy(src, dst)
}

func defaultClientHook(dst, src io.ReadWriteCloser) {
	Forward(dst, src)
}

func defaultServerHook(dst, src io.ReadWriteCloser) {
	defer src.Close()
	buf := make([]byte, 2048)
	var count int64
	for {
		n, err := src.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		count = count + int64(n)
		log.Printf("read bytes:%dB, Total: %dK %dM %dG", n, count>>10, count>>20, count>>30)
	}
}

func client(t *Tunnel) {
	log.Println("Client Start At ", t.listen.Addr().String())
	for {
		conn, err := t.listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		tConn, err := t.New()
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		if t.hook == nil {
			t.hook = defaultClientHook
		}
		go t.hook(tConn, conn)
	}
}

func server(t *Tunnel) {
	log.Println("Server Start At ", t.listen.Addr().String())
	for {
		conn, err := t.listen.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		tConn, err := createStream(conn)
		if err != nil {
			log.Println(err)
			conn.Close()
			continue
		}
		if t.hook == nil {
			t.hook = defaultServerHook
		}
		go t.hook(nil, tConn)
	}
}

package tunnel

import (
	"io"
	"net"
	"sync"
)

func Dial(addrPort string, key string, level int) (net.Conn, error) {
	conn, err := net.Dial("tcp", addrPort)
	if err != nil {
		return nil, err
	}
	sendAuth(conn, key)
	cmpConn, err := NewCompress(conn, level)
	if err != nil {
		conn.Close()
		return nil, err
	}
	chaConn, err := NewChacha20(cmpConn, key)
	if err != nil {
		cmpConn.Close()
		return nil, err
	}
	return chaConn, nil
}

func ListenAndServer(addrPort string, key string, level int, handleFunc func(net.Conn)) {
	listen, err := net.Listen("tcp", addrPort)
	if err != nil {
		panic(err)
	}
	println("tunnel listen on", addrPort)
	for {
		conn, err := listen.Accept()
		if err != nil {
			println(err.Error())
			continue
		}
		go handle(conn, key, level, handleFunc)
	}
}

func Forward(dst, src io.ReadWriteCloser) {
	var wg sync.WaitGroup
	forward := func(dst, src io.ReadWriteCloser) {
		defer src.Close()
		defer dst.Close()
		io.Copy(dst, src)
		wg.Done()
	}
	wg.Add(2)
	go forward(dst, src)
	go forward(src, dst)
	wg.Wait()
}

func handle(conn net.Conn, key string, level int, handleFunc func(net.Conn)) {
	if !auth(conn, key) {
		conn.Close()
		println("Auth False")
		return
	}
	cmpConn, err := NewCompress(conn, level)
	if err != nil {
		conn.Close()
		println(err)
		return
	}
	chaConn, err := NewChacha20(cmpConn, key)
	if err != nil {
		cmpConn.Close()
		println(err.Error())
		return
	}
	if handleFunc == nil {
		handleFunc = defaultHandle
	}
	handleFunc(chaConn)
}

func defaultHandle(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		println(buf[:n])
	}
}

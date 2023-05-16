package tunnel

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
)

func generateKey32(key string) []byte {
	// generate 32 bytes key
	return []byte(fmt.Sprintf("%x", md5.Sum([]byte(key))))
}

func sendAuth(conn io.Writer, key string) {
	kv := generateKey32(key)
	size := 8 + rand.Intn(8) // 前缀长度
	randData := make([]byte, size)
	rand.Read(randData)
	buf := bytes.NewBuffer([]byte{})
	buf.WriteByte(byte(size))
	buf.Write(randData)
	crc32Data := crc32.ChecksumIEEE(randData)
	binary.Write(buf, binary.BigEndian, crc32Data)
	conn.Write(xor(buf.Bytes(), kv)) // 建立加密连接前先发送随机前缀，避开特征检测
}

func xor(src, key []byte) []byte {
	if len(key) < len(src) {
		return src
	}
	for i := range src {
		src[i] = src[i] ^ key[i]
	}
	return src
}

func auth(conn io.Reader, key string) bool {
	kv := generateKey32(key)
	buf := make([]byte, 256)
	if _, err := conn.Read(buf[:1]); err != nil {
		println(err.Error())
		return false
	}
	i := buf[0] ^ kv[0]
	if _, err := conn.Read(buf[:i+4]); err != nil {
		println(err.Error())
		return false
	}
	encodes := xor(buf[:i+4], kv[1:])
	crc32Data := crc32.ChecksumIEEE(encodes[:i])
	if crc32Data != binary.BigEndian.Uint32(encodes[i:i+4]) {
		return false
	}
	return true
}

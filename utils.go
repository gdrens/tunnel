package tunnel

import (
	"crypto/md5"
	"fmt"
)

func generateKey32(key string) []byte {
	// generate 32 bytes key
	return []byte(fmt.Sprintf("%x", md5.Sum([]byte(key))))
}

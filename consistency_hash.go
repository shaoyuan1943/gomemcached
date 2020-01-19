package gomemcached

import (
	"crypto/md5"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
)

var (
	HashCRC32Table = crc32.MakeTable(crc32.IEEE)
)

func KetamaHash(key string, index uint32) []uint32 {
	var b strings.Builder
	b.WriteString(key)
	b.WriteString("#")
	b.WriteString(strconv.Itoa((int)(index)))

	digest := md5.Sum([]byte(b.String()))
	hashs := make([]uint32, 4)
	for i := 0; i < 4; i++ {
		hashs[i] = (uint32(digest[3+i*4]&0xFF) << 24) | (uint32(digest[2+i*4]&0xFF) << 16) | (uint32(digest[1+i*4]&0xFF) << 8) | uint32(digest[0+i*4]&0xFF)
	}

	return hashs
}

func MakeHash(key string) uint32 {
	hashKey := crc32.Checksum([]byte(key), HashCRC32Table)
	fmt.Printf("hash: %v\n", hashKey)
	return hashKey
}

type SortList []uint32

func (s SortList) Len() int {
	return len(s)
}

func (s SortList) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s SortList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

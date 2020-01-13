package gomemcached

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func CreateCluster() *cluster {
	addrs := []string{
		"10.11.10.91", "10.11.133.161", "10.11.64.2", "12.65.89.35", "56.39.87.65",
		"121.14.64.115", "89.56.87.12", "89.62.53.87", "192.168.0.1", "78.95.64.52",
	}

	cl := createCluster(addrs)
	return cl
}

func BenchmarkCluster_FindServerByKey(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	cl := CreateCluster()
	for i := 0; i < b.N; i++ {
		key := RandString(8)
		s := cl.FindServerByKey(key)
		if s == nil {
			b.Errorf("not found server, key: %v\n", key)
		}
	}
}

func TestCluster_FindServerByKey(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	cl := CreateCluster()
	hitMap := make(map[string]int)
	for i := 0; i < 3000000; i++ {
		key := RandString(8)
		s := cl.FindServerByKey(key)
		if s == nil {
			t.Errorf("not found server, key: %v\n", key)
		} else {
			if _, ok := hitMap[s.Addr]; !ok {
				hitMap[s.Addr] = 1
			} else {
				hitMap[s.Addr] += 1
			}
		}
	}

	for k, v := range hitMap {
		fmt.Printf("%v\t%v\n", k, v)
	}
}

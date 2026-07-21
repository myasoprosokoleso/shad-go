package main

import (
	"hash/fnv"
	"strings"
)

const keyLen = 10

var charsSample = buildCharsSample()

func generateKey(s string) string {
	var key strings.Builder
	key.Grow(keyLen)

	hash, base := getHash(s), uint64(len(charsSample))
	for range keyLen {
		key.WriteByte(charsSample[hash%base])
		hash /= base
	}
	return key.String()
}

func getHash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func buildCharsSample() []byte {
	sample := make([]byte, 0, 10+26*2)
	for i := byte('0'); i <= '9'; i++ {
		sample = append(sample, i)
	}
	for i := byte('A'); i <= 'Z'; i++ {
		sample = append(sample, i)
	}
	for i := byte('a'); i <= 'z'; i++ {
		sample = append(sample, i)
	}
	return sample
}

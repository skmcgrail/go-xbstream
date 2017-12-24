package xbstream

import "encoding/binary"

func int4store(int int) []byte {
	return intStore(int, 4)
}

func int8store(int int) []byte {
	return intStore(int, 8)
}

func intStore(int, size int) []byte {
	enc := make([]byte, size)
	binary.LittleEndian.PutUint32(enc, uint32(int))
	return enc
}

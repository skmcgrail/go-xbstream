package xbstream

import "encoding/binary"

func int4store(int int) []byte {
	enc := make([]byte, 4)
	binary.LittleEndian.PutUint32(enc, uint32(int))
	return enc
}

func int8store(int int) []byte {
	enc := make([]byte, 8)
	binary.LittleEndian.PutUint64(enc, uint64(int))
	return enc
}

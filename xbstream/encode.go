package xbstream

import "encoding/binary"

func int8store(int uint64) []byte {
	enc := make([]byte, 8)
	binary.LittleEndian.PutUint64(enc, int)
	return enc
}

func uint8korr(p []byte) uint64 {
	return binary.LittleEndian.Uint64(p[:8])
}

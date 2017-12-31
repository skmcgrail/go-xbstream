package xbstream

import (
	"errors"
	"io"
)

type ChunkFlag uint8
type ChunkType uint8

const (
	MinimumChunkSize             = 10 * 1024 * 1024
	MaxPathLength                = 512
	FlagChunkIgnorable ChunkFlag = 0x01
)

const (
	ChunkTypePayload = ChunkType('P')
	ChunkTypeEOF     = ChunkType('E')
	ChunkTypeUnknown = ChunkType(0)
)

var (
	chunkMagic      = []uint8("XBSTCK01")
	StreamReadError = errors.New("xbstream read error")
)

type Chunk struct {
	ChunkHeader
	io.Reader
}

type ChunkHeader struct {
	Magic     []uint8
	Flags     ChunkFlag
	Type      ChunkType
	PathLen   uint32
	Path      []uint8
	PayLen    uint64
	PayOffset uint64
	Checksum  uint32
}

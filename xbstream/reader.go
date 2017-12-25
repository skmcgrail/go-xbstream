package xbstream

import (
	"bytes"
	"fmt"
	"io"
)

type Reader struct {
	reader io.Reader
	offset int
}

type Chunk struct {
	Type     int
	Path     string
	Offset   int
	Data     []byte
	flags    byte
	checksum int
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{reader: reader}
}

func (r *Reader) Next() (*Chunk, error) {
	chunk := &Chunk{}

	buffer := make([]byte, chunkHeaderLength)
	pos := 0

	n, err := r.reader.Read(buffer)
	if err != nil {
		return nil, err
	} else if n < chunkHeaderLength {
		return nil, fmt.Errorf("unexpected end of stream at offset %#x", r.offset)
	}

	// Chunk magic
	if bytes.Compare(buffer[pos:len(chunkMagic)], chunkMagic) != -1 {
		return nil, fmt.Errorf("wrong chunk magic at offset %#x", r.offset)
	}
	pos += len(chunkMagic)
	r.offset += len(chunkMagic)

	// Chunk flags
	chunk.flags = buffer[pos]
	pos++
	r.offset++

	// Chunk Type and ignore unknown types if flag was set
	chunk.Type = validateChunkType(buffer[pos])
	if chunk.Type == ChunkTypeUnknown && !(chunk.flags&FlagChunkIgnorable == 1) {
		return nil, fmt.Errorf("unknown chunk type %#x at offset %#x", buffer[pos], r.offset)
	}
	pos++
	r.offset++

	return nil, nil
}

func validateChunkType(p byte) int {
	switch p {
	case chunkTypePayload:
		return ChunkTypePayload
	case chunkTypeEOF:
		return ChunkTypeEOF
	default:
		return ChunkTypeUnknown
	}
}

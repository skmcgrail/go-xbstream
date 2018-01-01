package xbstream

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"sync"
)

type Writer struct {
	mutex  sync.Mutex
	writer io.WriteCloser
}

type File struct {
	path   []byte
	writer *Writer
	chunk  []byte
	pos    int // current chunk slice position
	free   int // remaining chunk bytes
	offset int // current file offset
}

func NewWriter(writer io.WriteCloser) *Writer {
	return &Writer{sync.Mutex{}, writer}
}

func (w *Writer) Create(path string) (*File, error) {
	if len(path) > MaxPathLength {
		return nil, errors.New("max path length exceeded")
	}

	return &File{
		path:   []byte(path),
		writer: w,
		chunk:  make([]byte, MinimumChunkSize),
		free:   MinimumChunkSize,
	}, nil
}

func (w *Writer) Close() error {
	return w.writer.Close()
}

func (f *File) Write(p []byte) (int, error) {
	if len(p) < f.free {
		n := copy(f.chunk[f.pos:], p)
		f.pos += n
		f.free -= n

		return len(p), nil
	}

	if err := f.Flush(); err != nil {
		return 0, err
	}

	return len(p), f.writeChunk(p)
}

func (f *File) writeChunk(p []byte) error {
	var err error
	buffer := new(bytes.Buffer)
	chunk := new(Chunk)

	// Chunk Magic
	chunk.Magic = make([]uint8, len(chunkMagic))
	copy(chunk.Magic, chunkMagic)
	if err = binary.Write(buffer, binary.BigEndian, &chunk.Magic); err != nil {
		return err
	}

	// Chunk Flags
	chunk.Flags = 0
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.Flags); err != nil {
		return err
	}

	// Chunk Type
	chunk.Type = ChunkTypePayload
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.Type); err != nil {
		return err
	}

	// path Length
	chunk.PathLen = uint32(len(f.path))
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.PathLen); err != nil {
		return err
	}

	// path
	chunk.Path = f.path
	if err = binary.Write(buffer, binary.BigEndian, &chunk.Path); err != nil {
		return err
	}

	// Payload Length
	chunk.PayLen = uint64(len(p))
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.PayLen); err != nil {
		return err
	}

	// Checksum
	chunk.Checksum = crc32.ChecksumIEEE(p)

	f.writer.mutex.Lock()
	defer f.writer.mutex.Unlock()

	// Payload Offset
	chunk.PayOffset = uint64(f.offset)
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.PayOffset); err != nil {
		return err
	}

	// Checksum
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.Checksum); err != nil {
		return err
	}

	if _, err = io.Copy(f.writer.writer, buffer); err != nil {
		return nil
	}

	if _, err = io.Copy(f.writer.writer, bytes.NewReader(p)); err != nil {
		return err
	}

	f.offset += len(p)

	return nil
}

func (f *File) writeEOF() error {
	var err error
	buffer := new(bytes.Buffer)
	chunk := new(Chunk)

	f.writer.mutex.Lock()
	defer f.writer.mutex.Unlock()

	// Chunk Magic
	chunk.Magic = make([]uint8, len(chunkMagic))
	copy(chunk.Magic, chunkMagic)
	if err = binary.Write(buffer, binary.BigEndian, &chunk.Magic); err != nil {
		return err
	}

	// Chunk Flags
	chunk.Flags = 0
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.Flags); err != nil {
		return err
	}

	// Chunk Type
	chunk.Type = ChunkTypeEOF
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.Type); err != nil {
		return err
	}

	// path Length
	chunk.PathLen = uint32(len(f.path))
	if err = binary.Write(buffer, binary.LittleEndian, &chunk.PathLen); err != nil {
		return err
	}

	// path
	chunk.Path = f.path
	if err = binary.Write(buffer, binary.BigEndian, &chunk.Path); err != nil {
		return err
	}

	if _, err = io.Copy(f.writer.writer, buffer); err != nil {
		return err
	}

	return nil
}

func (f *File) Flush() error {
	if f.pos == 0 {
		return nil
	}

	if err := f.writeChunk(f.chunk[:f.pos]); err != nil {
		return err
	}

	f.pos = 0
	f.free = MinimumChunkSize

	return nil
}

func (f *File) Close() error {
	if err := f.Flush(); err != nil {
		return err
	}

	if err := f.writeEOF(); err != nil {
		return err
	}

	return nil
}

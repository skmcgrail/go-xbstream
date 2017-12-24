package xbstream

import (
	"bytes"
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
	magic := []byte(ChunkMagic)
	buffer := make([]byte, len(magic)-1+1+1+4+MaxPathLength+8+8+4)
	pos := 0
	n := 0

	// Chunk Magic
	n = copy(buffer[pos:], magic)
	pos += n

	// Chunk Flags
	buffer[pos] = 0
	pos++

	// Chunk Type
	n = copy(buffer[pos:], []byte(ChunkTypePayload))
	pos += n

	// path Length
	n = copy(buffer[pos:], int4store(len(f.path)))
	pos += n

	// path
	n = copy(buffer[pos:], f.path)
	pos += len(f.path)

	// Payload Length
	n = copy(buffer[pos:], int8store(len(p)))
	pos += n

	// Checksum
	checksum := crc32.ChecksumIEEE(p)

	f.writer.mutex.Lock()
	defer f.writer.mutex.Unlock()

	// Payload Offset
	n = copy(buffer[pos:], int8store(f.offset))
	pos += n

	n = copy(buffer[pos:], int4store(int(checksum)))
	pos += n

	_, err := io.Copy(f.writer.writer, bytes.NewReader(buffer))
	if err != nil {
		return err
	}

	_, err = io.Copy(f.writer.writer, bytes.NewReader(p))
	if err != nil {
		return err
	}

	f.offset += len(p)

	return nil
}

func (f *File) writeEOF() error {
	magic := []byte(ChunkMagic)
	buffer := make([]byte, len(magic)-1+1+1+4+MaxPathLength)
	pos := 0
	n := 0

	f.writer.mutex.Lock()
	defer f.writer.mutex.Unlock()

	// Chunk Magic
	n = copy(buffer[pos:], magic)
	pos += n

	// Chunk Flags
	buffer[pos] = 0
	pos += 1

	// Chunk Type
	n = copy(buffer[pos:], []byte(ChunkTypeEOF))
	pos += n

	// path Length
	n = copy(buffer[pos:], int4store(len(f.path)))
	pos += n

	// path
	n = copy(buffer[pos:], f.path)
	pos += len(f.path)

	_, err := io.Copy(f.writer.writer, bytes.NewReader(buffer))
	if err != nil {
		return err
	}

	return nil
}

func (f *File) Flush() error {
	if f.pos == 0 {
		return nil
	}

	if err := f.writeChunk(f.chunk); err != nil {
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

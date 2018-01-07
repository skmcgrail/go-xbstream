/*
 * Copyright (C) 2017 Sean McGrail
 * Copyright (C) 2011-2017 Percona LLC and/or its affiliates.
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
 */

package xbstream

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// Reader provides sequential access to chunks from an xbstream. Each chunk returned represents a
// contiguous set of bytes for a file stored in the xbstream archive. The Next method advances the stream
// and returns the next chunk in the archive. Each archive then acts as a reader for its contiguous set of bytes
type Reader struct {
	reader io.Reader
}

// NewReader creates a new Reader by wrapping the provided reader
func NewReader(reader io.Reader) *Reader {
	return &Reader{reader: reader}
}

// Next advances the Reader and returns the next Chunk.
// Note: end of input is represented by a specific Chunk type.
func (r *Reader) Next() (*Chunk, error) {
	var (
		chunk = new(Chunk)
		err   error
	)

	chunk.Magic = make([]uint8, len(chunkMagic))

	// Chunk Magic
	if err = binary.Read(r.reader, binary.BigEndian, &chunk.Magic); err != nil {
		// We should gracefully bubble up EOF if we attempt to read a new Chunk and hit EOF
		if err != io.EOF {
			return nil, ErrStreamRead
		}

		return nil, err
	}

	if bytes.Compare(chunk.Magic, chunkMagic) != 0 {
		return nil, errors.New("wrong chunk magic")
	}

	// Chunk Flags
	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.Flags); err != nil {
		return nil, ErrStreamRead
	}

	// Chunk Type
	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.Type); err != nil {
		return nil, ErrStreamRead
	}
	if chunk.Type = validateChunkType(chunk.Type); chunk.Type == ChunkTypeUnknown {
		if !(chunk.Flags&FlagChunkIgnorable == 1) {
			return nil, errors.New("unknown chunk type")
		}
	}

	// Path Length
	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.PathLen); err != nil {
		return nil, ErrStreamRead
	}

	// Path
	if chunk.PathLen > 0 {
		chunk.Path = make([]uint8, chunk.PathLen)
		if err = binary.Read(r.reader, binary.BigEndian, &chunk.Path); err != nil {
			return nil, ErrStreamRead
		}
	}

	if chunk.Type == ChunkTypeEOF {
		return chunk, nil
	}

	if binary.Read(r.reader, binary.LittleEndian, &chunk.PayLen); err != nil {
		return nil, ErrStreamRead
	}

	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.PayOffset); err != nil {
		return nil, ErrStreamRead
	}

	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.Checksum); err != nil {
		return nil, ErrStreamRead
	}

	if chunk.PayLen > 0 {
		buffer := bytes.NewBuffer(nil)
		if _, err := io.CopyN(buffer, r.reader, int64(chunk.PayLen)); err != nil {
			return nil, ErrStreamRead
		}
		chunk.Reader = buffer
	} else {
		chunk.Reader = bytes.NewReader(nil)
	}

	return chunk, nil
}

func validateChunkType(p ChunkType) ChunkType {
	switch p {
	case ChunkTypePayload:
		fallthrough
	case ChunkTypeEOF:
		return p
	default:
		return ChunkTypeUnknown
	}
}

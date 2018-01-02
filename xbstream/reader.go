/*
 * Copyright (C) 2017 Sean McGrail
 * Copyright (c) 2011-2017 Percona LLC and/or its affiliates.
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

type Reader struct {
	reader io.Reader
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{reader: reader}
}

func (r *Reader) Next() (*Chunk, error) {
	var (
		chunk = new(Chunk)
		err   error
	)

	chunk.Magic = make([]uint8, len(chunkMagic))

	// Chunk Magic
	if err = binary.Read(r.reader, binary.BigEndian, &chunk.Magic); err != nil {
		return nil, StreamReadError
	}

	if bytes.Compare(chunk.Magic, chunkMagic) != 0 {
		return nil, errors.New("wrong chunk magic")
	}

	// Chunk Flags
	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.Flags); err != nil {
		return nil, StreamReadError
	}

	// Chunk Type
	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.Type); err != nil {
		return nil, StreamReadError
	}
	if chunkType := validateChunkType(chunk.Type); chunkType == ChunkTypeUnknown {
		if !(chunk.Flags&FlagChunkIgnorable == 1) {
			return nil, errors.New("unknown chunk type")
		}
	}

	// Path Length
	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.PathLen); err != nil {
		return nil, StreamReadError
	}

	// Path
	if chunk.PathLen > 0 {
		chunk.Path = make([]uint8, chunk.PathLen)
		if err = binary.Read(r.reader, binary.BigEndian, &chunk.Path); err != nil {
			return nil, StreamReadError
		}
	}

	if chunk.Type == ChunkTypeEOF {
		return chunk, nil
	}

	if binary.Read(r.reader, binary.LittleEndian, &chunk.PayLen); err != nil {
		return nil, StreamReadError
	}

	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.PayOffset); err != nil {
		return nil, StreamReadError
	}

	if err = binary.Read(r.reader, binary.LittleEndian, &chunk.Checksum); err != nil {
		return nil, StreamReadError
	}

	if chunk.PayLen > 0 {
		chunk.Reader = io.LimitReader(r.reader, int64(chunk.PayLen))
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

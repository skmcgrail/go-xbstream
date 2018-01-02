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
	ChunkTypePayload = ChunkType('P') // Payload
	ChunkTypeEOF     = ChunkType('E') // End Of File
	ChunkTypeUnknown = ChunkType(0)   // Unknown Type
)

var (
	chunkMagic      = []uint8("XBSTCK01")
	StreamReadError = errors.New("xbstream read error")
)

// Chunk encapsulates a ChunkHeader and provides a io.Reader interface for reading the payload described by the Header
type Chunk struct {
	ChunkHeader
	io.Reader
}

// ChunkHeader contains the metadata regarding the payload that immediately follows within the archive
type ChunkHeader struct {
	Magic     []uint8
	Flags     ChunkFlag
	Type      ChunkType // The type of Chunk, Note xbstream archives end with a specific EOF type
	PathLen   uint32
	Path      []uint8
	PayLen    uint64
	PayOffset uint64
	Checksum  uint32
}

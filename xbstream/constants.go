package xbstream

const (
	MinimumChunkSize   = 10 * 1024 * 1024
	MaxPathLength      = 512
	FlagChunkIgnorable = 0x01
)

const (
	ChunkTypeUnknown = iota
	ChunkTypePayload
	ChunkTypeEOF
)

var (
	chunkHeaderLength = len(chunkMagic) - 1 + 1 + 1 + 4
	chunkMagic        = []byte("XBSTCK01")
	chunkTypePayload  = byte('P')
	chunkTypeEOF      = byte('E')
)

package xbstream

const (
	MaxPathLength    = 512
	MinimumChunkSize = 10 * 1024 * 1024
	ChunkMagic       = "XBSTCK01"
	ChunkTypeUnknown = "\x00"
	ChunkTypePayload = "P"
	ChunkTypeEOF     = "E"
)

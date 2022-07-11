package rtmp

type Chunk struct {
	BasicHeader   ChunkBasicHeader
	MessageHeader ChunkMsgHeader
	ExtTimestamp  uint32
	Payload       []byte
}

type ChunkBasicHeader struct {
	Format uint8
	Csid   uint32
}

type ChunkMsgHeader struct {
	Timestamp       uint32
	MessageLength   uint32
	MessageTypeId   uint8
	MessageStreamId uint32
}

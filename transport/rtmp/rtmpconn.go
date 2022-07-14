package rtmp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type Conn struct {
	Conn                net.Conn
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	Chunks              map[uint32]*Chunk
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
}

func (c *Conn) Handshake() error {
	var C0C1C2 = make([]byte, 1+1536*2)
	C0 := C0C1C2[0]
	C1 := C0C1C2[1 : 1536+1]
	C2 := C0C1C2[1536+1:]
	C0C1 := C0C1C2[:1536+1]

	var S0S1S2 = make([]byte, 1+1536*2)
	S0 := S0S1S2[0]
	S1 := S0S1S2[1 : 1536+1]
	S2 := S0S1S2[1536+1:]

	c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if _, err := io.ReadFull(c.Conn, C0C1); err != nil {
		return err
	}
	if C0 != 0x03 {
		return fmt.Errorf("rtmp: handshake: invalid version: %d, only support %d", C0, S0)
	}

	S0 = 0x03
	binary.BigEndian.PutUint32(S1[0:4], uint32(time.Now().Unix()))

	clientTime := binary.BigEndian.Uint32(C1[0:4])
	clientVersion := binary.BigEndian.Uint32(C1[4:8])
	if clientVersion == 0 {
		copy(S2, C1)
	} else {
		fmt.Print(clientTime)
		return fmt.Errorf("rtmp: handshake: not support crypte yet")
	}

	c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	if _, err := c.Conn.Write(S0S1S2); err != nil {
		return err
	}

	if _, err := io.ReadFull(c.Conn, C2); err != nil {
		return err
	}
	if !bytes.Equal(C2[8:], S1[8:]) {
		return fmt.Errorf("rtmp: handshake: C2 is not equal S1")
	}
	return nil
}

func (c *Conn) ReadChunks() (ret *Chunk, err error) {
	for {
		// read chunk
		// read chunk basic header

		basicHeaderBytes := make([]byte, 1)
		c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
		if _, err = io.ReadFull(c.Conn, basicHeaderBytes); err != nil {
			return
		}
		format := basicHeaderBytes[0] >> 6
		csid := uint32(basicHeaderBytes[0] & 0x3F)
		csidLen := 0
		switch csid {
		case 0:
			csidLen = 2
		case 1:
			csidLen = 3
		case 2:
			csidLen = 1
		}

		if csidLen-1 > 0 {
			c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
			csidBytes := make([]byte, csidLen-1)
			if _, err = io.ReadFull(c.Conn, csidBytes); err != nil {
				return
			}
			switch csidLen - 1 {
			case 1:
				csid = uint32(csidBytes[0]) + 64
			case 2:
				csid = uint32(csidBytes[1])*256 + uint32(csidBytes[0]) + 64
			}
		}
		var basicHeader = ChunkBasicHeader{
			CSID:   csid,
			Format: format,
		}

		// TODO read chunks from cached map and then fill the previous header into that.

		// read chunk msg header
		var msgHeaderLen int
		switch format {
		case 0:
			msgHeaderLen = 11
		case 1:
			msgHeaderLen = 7
		case 2:
			msgHeaderLen = 3
		case 3:
			msgHeaderLen = 0
		}

		// read basic message header
		var timestamp uint32
		var messageLen uint32
		var messageTypeId uint8
		var messageStreamId uint32
		if msgHeaderLen > 0 {
			c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
			msgHeaderBytes := make([]byte, msgHeaderLen)
			if _, err = io.ReadFull(c.Conn, msgHeaderBytes); err != nil {
				return
			}
			timestamp = uint32(msgHeaderBytes[2]) | uint32(msgHeaderBytes[1])<<8 | uint32(msgHeaderBytes[0])<<16
			if format == 0 || format == 1 {
				messageLen = uint32(msgHeaderBytes[5]) | uint32(msgHeaderBytes[4])<<8 | uint32(msgHeaderBytes[3])<<16
				messageTypeId = uint8(msgHeaderBytes[6])
			}
			if format == 0 {
				messageStreamId = uint32(msgHeaderBytes[10]) | uint32(msgHeaderBytes[9])<<8 | uint32(msgHeaderBytes[8])<<16 | uint32(msgHeaderBytes[7])
			}
		}

		var msgHeader = ChunkMsgHeader{
			Timestamp:       timestamp,
			MessageLength:   messageLen,
			MessageTypeId:   messageTypeId,
			MessageStreamId: messageStreamId,
		}

		var extTimestamp uint32
		if msgHeader.Timestamp == 0x00FFFFFF {
			c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
			extTimestampBytes := make([]byte, 4)
			if _, err = io.ReadFull(c.Conn, extTimestampBytes); err != nil {
				return
			}
			extTimestamp = binary.BigEndian.Uint32(extTimestampBytes)
		}

		chk, ok := c.Chunks[basicHeader.CSID]
		if !ok {
			c.Chunks[basicHeader.CSID] = &Chunk{
				BasicHeader:   basicHeader,
				MessageHeader: msgHeader,
				ExtTimestamp:  extTimestamp,
				Payload:       make([]byte, 0, msgHeader.MessageLength),
			}
		}
		var size uint32
		var remain uint32
		if remain = chk.MessageHeader.MessageLength - uint32(len(chk.Payload)); remain <= c.chunkSize {
			size = remain
		} else {
			size = c.chunkSize
		}
		c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
		buf := chk.Payload[len(chk.Payload) : uint32(len(chk.Payload))+size]
		if _, err = io.ReadFull(c.Conn, buf); err != nil {
			return
		}
		if uint32(len(chk.Payload)) == chk.MessageHeader.MessageLength {
			ret = chk
			return
		}
	}
}

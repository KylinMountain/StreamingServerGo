package rtmp

import (
	"encoding/binary"
	"github.com/rs/zerolog"
	"io"
	"net"
	"os"
	"time"
)

type RTMP struct {
	logger       zerolog.Logger
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewRTMPServer() *RTMP {
	return &RTMP{
		logger:       zerolog.New(os.Stdout),
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
}

func (r *RTMP) Serve(listener net.Listener) {
	for {
		netConn, err := listener.Accept()
		if err != nil {
			continue
		}
		conn := &Conn{
			Conn:         netConn,
			ReadTimeout:  r.ReadTimeout,
			WriteTimeout: r.WriteTimeout,
		}
		go r.HandleMsg(conn)
	}
}

func (r *RTMP) HandleMsg(conn *Conn) error {
	err := conn.Handshake()
	if err != nil {
		return err
	}

	conn.Conn.SetReadDeadline(time.Now().Add(conn.ReadTimeout))
	// read chunk
	// read chunk basic header
	basicHeaderBytes := make([]byte, 1)
	if _, err = io.ReadFull(conn.Conn, basicHeaderBytes); err != nil {
		return err
	}
	fmt := basicHeaderBytes[0] >> 6
	csid := uint32(basicHeaderBytes[0] & 0x3F)
	csidLen := 0
	switch csid {
	case 0:
		csidLen = 2
	case 1:
		csidLen = 3
	case 2:
		csidLen = 1
	default:
		csidLen = 1
	}

	if csidLen-1 > 0 {
		conn.Conn.SetReadDeadline(time.Now().Add(conn.ReadTimeout))
		csidBytes := make([]byte, csidLen-1)
		if _, err = io.ReadFull(conn.Conn, csidBytes); err != nil {
			return err
		}
		switch csidLen - 1 {
		case 1:
			csid = uint32(csidBytes[0]) + 64
		case 2:
			csid = uint32(csidBytes[1])*256 + uint32(csidBytes[0]) + 64
		}
	}
	var basicHeader = ChunkBasicHeader{
		Csid:   csid,
		Format: fmt,
	}

	// read chunk msg header
	var msgHeaderLen int
	switch fmt {
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
		conn.Conn.SetReadDeadline(time.Now().Add(conn.ReadTimeout))
		msgHeaderBytes := make([]byte, msgHeaderLen)
		if _, err = io.ReadFull(conn.Conn, msgHeaderBytes); err != nil {
			return err
		}
		timestamp = uint32(msgHeaderBytes[2]) | uint32(msgHeaderBytes[1])<<8 | uint32(msgHeaderBytes[0])<<16
		if fmt == 0 || fmt == 1 {
			messageLen = uint32(msgHeaderBytes[5]) | uint32(msgHeaderBytes[4])<<8 | uint32(msgHeaderBytes[3])<<16
			messageTypeId = uint8(msgHeaderBytes[6])
		}
		if fmt == 0 {
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
		conn.Conn.SetReadDeadline(time.Now().Add(conn.ReadTimeout))
		extTimestampBytes := make([]byte, 4)
		if _, err = io.ReadFull(conn.Conn, extTimestampBytes); err != nil {
			return err
		}
		extTimestamp = binary.BigEndian.Uint32(extTimestampBytes)
	}

	conn.Conn.SetReadDeadline(time.Now().Add(conn.ReadTimeout))
	payload := make([]byte, msgHeader.MessageLength)
	if _, err = io.ReadFull(conn.Conn, payload); err != nil {
		return err
	}

	var chunk = Chunk{
		BasicHeader:   basicHeader,
		MessageHeader: msgHeader,
		ExtTimestamp:  extTimestamp,
		Payload:       payload,
	}

	switch chunk.MessageHeader.MessageTypeId {
	case fmt:
		
	}

	return nil
}

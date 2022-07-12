package rtmp

import (
	"github.com/rs/zerolog"
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
			Conn:            netConn,
			ReadTimeout:     r.ReadTimeout,
			WriteTimeout:    r.WriteTimeout,
			chunkSize:       128,
			remoteChunkSize: 128,
		}
		go r.HandleMsg(conn)
	}
}

func (r *RTMP) HandleMsg(conn *Conn) error {
	err := conn.Handshake()
	if err != nil {
		return err
	}
	chunk, err := conn.ReadChunks()

	switch chunk.MessageHeader.MessageTypeId {
	case chunk.BasicHeader.Format:

	}

	return nil
}

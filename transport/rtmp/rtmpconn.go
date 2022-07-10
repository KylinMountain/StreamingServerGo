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
	Conn         net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
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
	if !bytes.Equal(C2, S1) {
		return fmt.Errorf("rtmp: handshake: C2 is not equal S1")
	}
	return nil
}

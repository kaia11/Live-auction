package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	websocketGUID     = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	opcodeText        = 0x1
	opcodeClose       = 0x8
	opcodePing        = 0x9
	opcodePong        = 0xA
	readWait          = 70 * time.Second
	writeWait         = 10 * time.Second
	pingInterval      = 25 * time.Second
	maxControlPayload = 125
)

type WSConn struct {
	conn net.Conn
	mu   sync.Mutex
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*WSConn, error) {
	if !headerHasToken(r.Header, "Connection", "upgrade") {
		return nil, errors.New("missing Connection upgrade header")
	}
	if !strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		return nil, errors.New("missing Upgrade websocket header")
	}
	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		return nil, errors.New("unsupported websocket version")
	}

	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		return nil, errors.New("missing websocket key")
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("response writer does not support hijacking")
	}

	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}

	accept := computeAcceptKey(key)
	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n",
		accept,
	)

	if _, err := rw.WriteString(response); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := rw.Flush(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &WSConn{
		conn: conn,
	}, nil
}

func (c *WSConn) Close() error {
	return c.conn.Close()
}

func (c *WSConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *WSConn) WriteJSON(v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.writeFrame(opcodeText, payload)
}

func (c *WSConn) WritePing() error {
	return c.writeFrame(opcodePing, nil)
}

func (c *WSConn) WritePong(payload []byte) error {
	if len(payload) > maxControlPayload {
		payload = payload[:maxControlPayload]
	}
	return c.writeFrame(opcodePong, payload)
}

func (c *WSConn) WriteClose() error {
	return c.writeFrame(opcodeClose, nil)
}

func (c *WSConn) ReadLoop(onPong func(), onPing func([]byte) error) error {
	reader := bufio.NewReader(c.conn)

	for {
		opcode, payload, err := readFrame(reader)
		if err != nil {
			return err
		}

		switch opcode {
		case opcodeClose:
			_ = c.WriteClose()
			return io.EOF
		case opcodePing:
			if onPing != nil {
				if err := onPing(payload); err != nil {
					return err
				}
			}
		case opcodePong:
			if onPong != nil {
				onPong()
			}
		default:
			// The current server does not consume client text/binary data.
		}
	}
}

func (c *WSConn) writeFrame(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	header := []byte{0x80 | opcode}
	length := len(payload)
	switch {
	case length < 126:
		header = append(header, byte(length))
	case length <= 65535:
		header = append(header, 126, byte(length>>8), byte(length))
	default:
		header = append(header, 127)
		sizeBuf := make([]byte, 8)
		binary.BigEndian.PutUint64(sizeBuf, uint64(length))
		header = append(header, sizeBuf...)
	}

	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	if length == 0 {
		return nil
	}
	_, err := c.conn.Write(payload)
	return err
}

func readFrame(reader *bufio.Reader) (byte, []byte, error) {
	first, err := reader.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	second, err := reader.ReadByte()
	if err != nil {
		return 0, nil, err
	}

	opcode := first & 0x0F
	masked := second&0x80 != 0
	payloadLen := int(second & 0x7F)

	switch payloadLen {
	case 126:
		extended := make([]byte, 2)
		if _, err := io.ReadFull(reader, extended); err != nil {
			return 0, nil, err
		}
		payloadLen = int(binary.BigEndian.Uint16(extended))
	case 127:
		extended := make([]byte, 8)
		if _, err := io.ReadFull(reader, extended); err != nil {
			return 0, nil, err
		}
		payloadLen64 := binary.BigEndian.Uint64(extended)
		if payloadLen64 > 1<<31-1 {
			return 0, nil, errors.New("frame too large")
		}
		payloadLen = int(payloadLen64)
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(reader, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return 0, nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

func computeAcceptKey(key string) string {
	sum := sha1.Sum([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func headerHasToken(header http.Header, key string, want string) bool {
	for _, value := range header.Values(key) {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			if strings.EqualFold(strings.TrimSpace(part), want) {
				return true
			}
		}
	}
	return false
}

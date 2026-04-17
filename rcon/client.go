package rcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	PacketAuth    = 3
	PacketCommand = 2
	PacketResp    = 0

	// maxSkipPackets is how many unrelated packets we tolerate before giving up.
	maxSkipPackets = 32
)

type Client struct {
	conn    net.Conn
	mu      sync.Mutex
	timeout time.Duration
	counter int32
}

func New(addr string, password string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	c := &Client{conn: conn, timeout: timeout}
	if err := c.authenticate(password); err != nil {
		conn.Close()
		return nil, err
	}
	return c, nil
}

func (c *Client) nextID() int32 {
	for {
		id := atomic.AddInt32(&c.counter, 1)
		if id > 0 {
			return id
		}
	}
}

func (c *Client) authenticate(password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sentID, err := c.send(PacketAuth, password)
	if err != nil {
		return err
	}

	for skip := 0; skip < maxSkipPackets; skip++ {
		respID, _, _, err := c.readPacket()
		if err != nil {
			return err
		}
		if respID == -1 {
			return fmt.Errorf("authentication failed: wrong password")
		}
		if respID == sentID {
			return nil
		}
	}
	return fmt.Errorf("authentication failed: no matching response after %d packets", maxSkipPackets)
}

func (c *Client) Execute(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sentID, err := c.send(PacketCommand, cmd)
	if err != nil {
		return "", err
	}

	for skip := 0; skip < maxSkipPackets; skip++ {
		respID, _, body, err := c.readPacket()
		if err != nil {
			return "", err
		}
		if respID == sentID {
			return body, nil
		}
	}
	return "", fmt.Errorf("no matching response after %d packets", maxSkipPackets)
}

func (c *Client) send(typ int32, body string) (int32, error) {
	_ = c.conn.SetWriteDeadline(time.Now().Add(c.timeout))

	id := c.nextID()
	buf := new(bytes.Buffer)
	length := int32(len(body) + 10)

	_ = binary.Write(buf, binary.LittleEndian, length)
	_ = binary.Write(buf, binary.LittleEndian, id)
	_ = binary.Write(buf, binary.LittleEndian, typ)
	buf.WriteString(body)
	buf.Write([]byte{0x00, 0x00})

	if _, err := c.conn.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return id, nil
}

func (c *Client) readPacket() (int32, int32, string, error) {
	_ = c.conn.SetReadDeadline(time.Now().Add(c.timeout))

	var length int32
	if err := binary.Read(c.conn, binary.LittleEndian, &length); err != nil {
		return 0, 0, "", err
	}
	if length < 10 || length > 4096 {
		return 0, 0, "", fmt.Errorf("invalid packet length: %d", length)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return 0, 0, "", err
	}

	reader := bytes.NewReader(data)
	var id, typ int32
	_ = binary.Read(reader, binary.LittleEndian, &id)
	_ = binary.Read(reader, binary.LittleEndian, &typ)

	bodyLen := length - 10
	if bodyLen <= 0 {
		return id, typ, "", nil
	}
	body := make([]byte, bodyLen)
	_, _ = reader.Read(body)

	return id, typ, string(body), nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

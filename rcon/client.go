package rcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	PacketAuth    = 3
	PacketCommand = 2
)

type Client struct {
	conn net.Conn
	mu   sync.Mutex
}

// New connects to a Minecraft RCON server
func New(addr string, password string, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	c := &Client{conn: conn}
	if err := c.authenticate(password); err != nil {
		conn.Close()
		return nil, err
	}
	return c, nil
}

func (c *Client) authenticate(password string) error {
	resp, err := c.send(PacketAuth, password)
	if err != nil {
		return err
	}
	if resp == -1 {
		return fmt.Errorf("authentication failed")
	}
	return nil
}

// Execute sends a command and returns the response string
func (c *Client) Execute(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.send(PacketCommand, cmd)
	if err != nil {
		return "", err
	}

	// Read response
	var length int32
	if err := binary.Read(c.conn, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return "", err
	}

	// Skip ID (4) and Type (4), then read body until null terminator
	return string(data[8 : length-2]), nil
}

func (c *Client) send(typ int32, body string) (int32, error) {
	id := int32(time.Now().UnixNano() / 1000000)
	buf := new(bytes.Buffer)

	_ = binary.Write(buf, binary.LittleEndian, int32(len(body)+10)) // Length
	_ = binary.Write(buf, binary.LittleEndian, id)                 // ID
	_ = binary.Write(buf, binary.LittleEndian, typ)                // Type
	buf.WriteString(body)                                          // Body
	buf.Write([]byte{0x00, 0x00})                                  // Terminators

	if _, err := c.conn.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return id, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

package protocol

import (
	"errors"
	"fmt"
)

type PacketReader struct {
	data   []byte
	offset int
}

func NewPacketReader(data []byte) *PacketReader {
	return &PacketReader{data: data, offset: 0}
}

func (pr *PacketReader) ReadVarInt() (int32, error) {
	var value int32
	var numRead int
	for {
		if numRead >= 5 {
			return 0, errors.New("varint too big")
		}

		if pr.offset >= len(pr.data) {
			return 0, errors.New("unexpected end of data")
		}

		b := pr.data[pr.offset]
		pr.offset++

		value |= int32(b&0x7F) << uint(7*numRead)
		numRead++

		if b&0x80 == 0 {
			break
		}
	}
	return value, nil
}

func (pr *PacketReader) ReadString() (string, error) {
	length, err := pr.ReadVarInt()
	if err != nil {
		return "", err
	}
	if pr.offset+int(length) > len(pr.data) {
		return "", errors.New("string length exceeds data")
	}
	str := string(pr.data[pr.offset : pr.offset+int(length)])
	pr.offset += int(length)
	return str, nil
}

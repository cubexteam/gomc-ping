package protocol

import (
	"bytes"
	"encoding/binary"
)

type PacketBuffer struct {
	buf *bytes.Buffer
}

func NewPacketBuffer() *PacketBuffer {
	return &PacketBuffer{buf: new(bytes.Buffer)}
}

func (pb *PacketBuffer) WriteVarInt(v int) {
	pb.buf.Write(WriteVarInt(v))
}

func (pb *PacketBuffer) WriteString(s string) {
	pb.WriteVarInt(len(s))
	pb.buf.WriteString(s)
}

func (pb *PacketBuffer) WriteUint16(v uint16) {
	var tmp [2]byte
	binary.BigEndian.PutUint16(tmp[:], v)
	pb.buf.Write(tmp[:])
}

func (pb *PacketBuffer) WriteUint64(v uint64) {
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], v)
	pb.buf.Write(tmp[:])
}

func (pb *PacketBuffer) Build() []byte {
	data := pb.buf.Bytes()
	var framed bytes.Buffer

	// Write length using centralized WriteVarInt
	framed.Write(WriteVarInt(len(data)))
	framed.Write(data)
	return framed.Bytes()
}

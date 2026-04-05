package protocol

import (
	"errors"
	"io"
)

// ReadVarInt reads a variable-length integer from a ByteReader
func ReadVarInt(r io.ByteReader) (int, error) {
	var value int
	var position int

	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}

		value |= int(b&0x7F) << (7 * position)
		position++

		if (b & 0x80) == 0 {
			break
		}

		if position >= 5 {
			return 0, errors.New("VarInt is too big")
		}
	}

	return value, nil
}

// ReadVarIntFromIO reads a VarInt directly from an io.Reader
func ReadVarIntFromIO(r io.Reader) (int, error) {
	var result int
	var numRead int
	var one [1]byte

	for {
		if _, err := io.ReadFull(r, one[:]); err != nil {
			return 0, err
		}
		b := one[0]
		result |= int(b&0x7F) << (7 * numRead)
		numRead++

		if (b & 0x80) == 0 {
			break
		}

		if numRead >= 5 {
			return 0, errors.New("VarInt is too big")
		}
	}
	return result, nil
}

// WriteVarInt encodes an integer into a VarInt byte slice
func WriteVarInt(value int) []byte {
	var buf []byte
	var v uint32 = uint32(value)
	for {
		if (v & ^uint32(0x7F)) == 0 {
			buf = append(buf, byte(v))
			return buf
		}

		buf = append(buf, byte((v&0x7F)|0x80))
		v >>= 7
	}
}

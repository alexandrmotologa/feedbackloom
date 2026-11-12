package store

import (
	"encoding/binary"
	"math"
)

// EncodeFloat32Slice converts a slice of float32 into a byte slice.
func EncodeFloat32Slice(v []float32) []byte {
	if len(v) == 0 {
		return nil
	}
	buf := make([]byte, len(v)*4)
	for i, val := range v {
		bits := math.Float32bits(val)
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], bits)
	}
	return buf
}

// DecodeFloat32Slice converts a byte slice into a slice of float32.
func DecodeFloat32Slice(b []byte) []float32 {
	if len(b) == 0 {
		return nil
	}
	n := len(b) / 4
	res := make([]float32, n)
	for i := 0; i < n; i++ {
		bits := binary.LittleEndian.Uint32(b[i*4 : (i+1)*4])
		res[i] = math.Float32frombits(bits)
	}
	return res
}

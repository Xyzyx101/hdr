package format

import (
	"encoding/binary"
	"math"
)

// ToBytes converts given float64 values to their bytes representation.
func ToBytes(f1, f2, f3 float64) []byte {
	pixel := make([]byte, 0, 3*4)

	pixel = append(pixel, float32bytes(float32(f1))...)
	pixel = append(pixel, float32bytes(float32(f2))...)
	pixel = append(pixel, float32bytes(float32(f3))...)

	return pixel
}

func float32bytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	return bytes
}

// FromBytes converts given bytes to their float64 values.
func FromBytes(pixel []byte) (float64, float64, float64) {
	f1 := float32frombytes(pixel[0:4])
	f2 := float32frombytes(pixel[4:8])
	f3 := float32frombytes(pixel[8:12])

	return float64(f1), float64(f2), float64(f3)
}

func float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)

	return float
}

package codec

type FixedUint64 [8]byte

func EncodeFixedUint64(n uint64) FixedUint64 {
	res := new(FixedUint64)
	res[0] = byte(n)
	res[1] = byte(n >> 8)
	res[2] = byte(n >> 16)
	res[3] = byte(n >> 24)
	res[4] = byte(n >> 32)
	res[5] = byte(n >> 40)
	res[6] = byte(n >> 48)
	res[7] = byte(n >> 56)
	return *res
}

func DecodeFixedUint64(n FixedUint64) uint64 {
	res := uint64(n[0]) + (uint64(n[1]) << 8) +
		(uint64(n[2]) << 16) + (uint64(n[3]) << 24) +
		(uint64(n[4]) << 32) + (uint64(n[5]) << 40) +
		(uint64(n[6]) << 48) + (uint64(n[7]) << 56)
	return res
}

func GetFixedUint64(data []byte, index int) FixedUint64 {
	res := new(FixedUint64)
	res[0] = data[index]
	res[1] = data[index+1]
	res[2] = data[index+2]
	res[3] = data[index+3]
	res[4] = data[index+4]
	res[5] = data[index+5]
	res[6] = data[index+6]
	res[7] = data[index+7]
	return *res
}

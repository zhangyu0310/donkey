package codec

import "errors"

type VarUint64 []byte

func EncodeVarUint64(n uint64) VarUint64 {
	res := make([]byte, 0, 1)
	for n >= 128 {
		res = append(res, byte(n|128))
		n >>= 7
	}
	res = append(res, byte(n))
	return res
}

func DecodeVarUint64(n VarUint64) uint64 {
	var res uint64
	index := 0
	for shift := 0; shift <= 63; shift += 7 {
		b := n[index]
		index++
		if (b & 128) != 0 {
			res |= uint64(b) & 127 << shift
		} else {
			res |= uint64(b) << shift
			break
		}
	}
	return res
}

func GetVarUint64(data []byte, index int) (VarUint64, int, error) {
	varUint64 := make(VarUint64, 0, 4)
	for {
		if len(data) == index {
			return nil, index, errors.New("data not enough")
		}
		varUint64 = append(varUint64, data[index])
		if data[index]&128 == 0 {
			index++
			break
		}
		index++
	}
	return varUint64, index, nil
}

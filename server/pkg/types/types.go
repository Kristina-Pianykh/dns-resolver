package types

type (
	Bit1  uint8
	Bit2  uint8
	Bit3  uint8
	Bit4  uint8
	Byte2 uint16
)

func ToBit1(b []byte) Bit1 {
	v := Bit1((b[0] >> 7) & 0b0000_0001)
	for idx := range b {
		b[idx] = b[idx] << 1
	}
	return v
}

func ToBit2(b []byte) Bit2 {
	v := Bit2(b[0] >> 6 & 0b0000_0011)
	for idx := range b {
		b[idx] = b[idx] << 2
	}
	return v
}

func ToBit3(b []byte) Bit3 {
	v := Bit3(b[0] >> 5 & 0b0000_0111)
	for idx := range b {
		b[idx] = b[idx] << 3
	}
	return v
}

func ToBit4(b []byte) Bit4 {
	v := Bit4(b[0] >> 4 & 0b0000_1111)
	for idx := range b {
		b[idx] = b[idx] << 4
	}
	return v
}

func ToByte2(b *[]byte) Byte2 {
	v := uint16((*b)[0])<<8 | uint16((*b)[1])
	*b = (*b)[2:]
	return Byte2(v)
}

package grpc

import "encoding/binary"

func ByteArray2Int64(buffer []byte) int64 {
	num, _ := binary.Varint(buffer)
	return num
}

func Int64ToByteArray(num int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, num)
	return buf[:n]
}

package sfs

import "encoding/binary"

func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// todo
func GenToken() string {
	return "abcdefgh"
}

// todo
func CheckUser(username, password string) bool {
	return true
}

func PutUint32(a uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, a)
	return b
}

func PutUint64(a uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, a)
	return b
}

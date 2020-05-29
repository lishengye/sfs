package sfs

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

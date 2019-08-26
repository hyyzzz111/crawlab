package utils

func EncryptPassword(str string) string {
	return MD5(str)
}

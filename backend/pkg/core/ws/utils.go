package ws

import (
	"io"
	"strings"
)

func Close(c io.Closer) {
	_ = c.Close()
}
func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := strings.Join([]string{absolutePath, relativePath}, ".")
	finalPath = strings.TrimLeft(finalPath, ".")
	appendSlash := lastChar(relativePath) == '.' && lastChar(finalPath) != '.'
	if appendSlash {
		return finalPath + "."
	}
	return finalPath
}
func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

package utils

import (
	"strings"
	"unicode"
)

// MixCase чередует верхний и нижний регистр букв строки
func MixCase(value string) string {
	var result strings.Builder
	result.Grow(len(value))

	upper := true
	for _, char := range value {
		if upper {
			result.WriteRune(unicode.ToUpper(char))
		} else {
			result.WriteRune(unicode.ToLower(char))
		}
		upper = !upper
	}
	return result.String()
}

package utils

import (
	"strings"
	"unicode"
)

// 檢查字串是否有中文
func CheckChineseExist(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// 不管有字串與字串之間有多少個空字元都用一個逗號隔開
func ReplaceSpaceToComma(str string) string {
	words := []string{}
	for _, word := range strings.Split(str, " ") {
		if len(word) > 0 {
			words = append(words, word)
		}
	}
	return strings.Join(words, ",")
}

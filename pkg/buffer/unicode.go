package buffer

import (
	"unicode/utf8"
)

// RuneIndex 计算字符串中指定rune索引对应的字节位置
func RuneIndex(s string, runeIndex int) int {
	if runeIndex < 0 {
		return -1
	}

	count := 0
	byteIndex := 0

	for count < runeIndex && byteIndex < len(s) {
		_, size := utf8.DecodeRuneInString(s[byteIndex:])
		byteIndex += size
		count++
	}

	if count < runeIndex {
		return -1 // 越界
	}

	return byteIndex
}

// RuneIndexRange 将rune索引范围转换为字节索引范围
func RuneIndexRange(s string, runeStart, runeEnd int) (byteStart, byteEnd int) {
	if runeStart < 0 || runeEnd < runeStart {
		return -1, -1
	}

	byteStart = RuneIndex(s, runeStart)
	if byteStart < 0 {
		return -1, -1
	}

	if runeStart == runeEnd {
		return byteStart, byteStart
	}

	// 计算结束位置
	count := runeStart
	byteIndex := byteStart

	for count < runeEnd && byteIndex < len(s) {
		_, size := utf8.DecodeRuneInString(s[byteIndex:])
		byteIndex += size
		count++
	}

	return byteStart, byteIndex
}

// RuneCount 返回字符串中Unicode字符(rune)的数量
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}

// ValidUTF8 检查字符串是否为有效的UTF-8编码
func ValidUTF8(s string) bool {
	return utf8.Valid([]byte(s))
}

// EnsureValidUTF8 确保返回的字符串是有效的UTF-8编码，如有无效字符则替换为Unicode替换字符
func EnsureValidUTF8(s string) string {
	if ValidUTF8(s) {
		return s
	}

	// 替换无效的UTF-8序列
	result := make([]rune, 0, len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			// 跳过无效的字节
			i++
		} else {
			result = append(result, r)
			i += size
		}
	}

	return string(result)
}

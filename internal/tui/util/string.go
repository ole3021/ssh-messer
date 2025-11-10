package util

// TruncateString 截断字符串到指定长度，如果超过长度则添加省略号
// maxLen: 最大长度（包括省略号）
// 如果 maxLen <= 0，返回原字符串
// 如果 maxLen <= 3，直接截断不添加省略号
// 否则截断并在末尾添加 "..."
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return string(runes[:maxLen-3]) + "..."
	}
	return string(runes[:maxLen])
}


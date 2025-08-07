package tools

import (
	"fmt"
	"strings"
)

// FormatLine 函数用于格式化输出行，使其总长度达到指定值
func FormatLine(prefix string, fillChar string) string {
	// 计算前缀字符串的长度
	prefixLength := len(prefix)
	// 计算需要补充的字符数量
	fillLength := 79 - prefixLength
	// 如果需要补充的字符数量小于等于 0，则直接返回前缀字符串
	if fillLength <= 0 {
		return prefix
	}
	// 生成补充字符的字符串
	fillStr := strings.Repeat(fillChar, fillLength)
	// 将前缀字符串和补充字符的字符串拼接起来
	return fmt.Sprintf("%s%s\n", prefix, fillStr)
}

// func TaskReport(results <-chan task.TaskStatus) string {
// 	var total, success, failed int
// 	for result := range results {
// 		total++
// 		switch result.Status {
// 		case "success":
// 			success++
// 		case "failure":
// 			failed++
// 		}
// 	}
// 	return fmt.Sprintf("Total: %d\t  Success:%d\t  Failed:%d\n", total, success, failed)
// }

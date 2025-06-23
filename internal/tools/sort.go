package tools

import (
	"fmt"
	"sort"
	"strings"
)

// 配置更新接口的hosts格式为server-ip，因此需要将数据归类存储
func SortHosts(data string) (map[string][]string, error) {
	// 创建一个map来存储数据
	prefixMap := make(map[string][]string)

	// 分割数据
	items := strings.Split(data, ",")
	for _, item := range items {
		// 分割前缀和IP
		parts := strings.Split(item, "-")
		if len(parts) == 2 {
			prefix := parts[0]
			ip := parts[1]
			// 将IP添加到对应的前缀键中
			prefixMap[prefix] = append(prefixMap[prefix], ip)
		} else {
			return nil, fmt.Errorf("invalid item format: %s", item)
		}
	}
	return prefixMap, nil
}
func ContainsKeysSorted(keys []string, target string) bool {
	index := sort.SearchStrings(keys, target)
	return index < len(keys) && keys[index] == target
}

package tools

import "fmt"

func AddNamespace(key string, ns string) string {
	return fmt.Sprintf("%s/%s", ns, key)
}
func RemoveNamespace(key string, ns string) string {
	return key[len(ns)+1:]
}

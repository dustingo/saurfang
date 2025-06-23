package tools

import (
	"fmt"
	"os"
)

func PathInit(paths []string) error {
	for _, p := range paths {
		_, err := os.Stat(p)
		if os.IsNotExist(err) {
			err = os.MkdirAll(p, 0755)
			if err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
		}
	}
	return nil
}

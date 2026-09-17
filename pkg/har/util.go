package har

import "os"

// readFileOS 操作系统读取
func readFileOS(path string) ([]byte, error) {
	return os.ReadFile(path)
}
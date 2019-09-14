package disk


import (
	"os"
)

func ExistsPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CreateFile(filePath string, content []byte) error {
	fc, err := os.Create(filePath)
	defer fc.Close()
	if err == nil {
		fc.Write(content)
	}
	return err
}
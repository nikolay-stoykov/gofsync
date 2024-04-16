package hash

import (
	"os"
)

type FileWriter struct {
	path string
}

// NewFileWriter creates a new FileWriter with the given file path.
func NewFileWriter(path string) *FileWriter {
	return &FileWriter{
		path: path,
	}
}

// Append writes the given bytes to the end of the file.
func (fw *FileWriter) Append(data []byte) error {
	file, err := os.OpenFile(fw.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

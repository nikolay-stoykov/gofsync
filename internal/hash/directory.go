package hash

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PathFileProcessor interface {
	Process(path string) error
}

type DirectoryHasher struct {
	metadataDirPath string
	path            string
	fileProcessor   PathFileProcessor
}

// NewDirectoryHasher creates a new FileReader with the given directory path and FileHasher.
func NewDirectoryHasher(path, metadataFilesDir string, fileProcessor PathFileProcessor) *DirectoryHasher {
	return &DirectoryHasher{
		path:            path,
		fileProcessor:   fileProcessor,
		metadataDirPath: metadataFilesDir,
	}
}

// ReadFiles reads all files in the directory and returns their contents as a map,
// with the SHA-256 hash of each file as the value.
func (fr *DirectoryHasher) ReadFiles() error {
	if err := fr.ensureMetadataDir(); err != nil {
		return err
	}

	start := time.Now()

	err := filepath.Walk(fr.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk errored: %w", err)
		}

		if strings.Contains(filepath.Dir(path), fr.metadataDirPath) {
			return nil
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		return fr.fileProcessor.Process(path)
	})

	if err != nil {
		return err
	}

	elapsed := time.Since(start)
	log.Println("Rehasher finished in", elapsed.Milliseconds())

	return nil
}

func (fr *DirectoryHasher) ensureMetadataDir() error {
	_, err := os.Stat(fr.metadataDirPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(fr.metadataDirPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

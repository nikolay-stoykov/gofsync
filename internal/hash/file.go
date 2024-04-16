package hash

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type ReaderHasher interface {
	Hash(io.Reader) ([]byte, bool, error)
}

type FileProcessor struct {
	hasher           ReaderHasher
	metadataFilesDir string
}

func NewFileProcessor(metadataFileDir string, hasher ReaderHasher) *FileProcessor {
	return &FileProcessor{
		metadataFilesDir: metadataFileDir,
		hasher:           hasher,
	}
}

func (f *FileProcessor) Process(path string) error {
	fReader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fReader.Close()

	fhash := sha256.Sum256([]byte(path))
	metadataFileName := Sha256ToString(fhash[:])

	metadataFilePath := filepath.Join(f.metadataFilesDir, string(metadataFileName))
	metadataFile, err := os.OpenFile(metadataFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not open metadata file: %w", err)
	}

	defer metadataFile.Close()

	for {
		hash, more, err := f.hasher.Hash(fReader)
		if err != nil {
			return fmt.Errorf("could not generate hash: %w", err)
		}
		if len(hash) == 0 {
			break
		}

		metadataBlock := make([]byte, 32)
		read, err := metadataFile.Read(metadataBlock)
		if err != nil && err != io.EOF {
			return err
		}

		if read > 0 && read < 32 {
			return fmt.Errorf("metadata file contains invalid block with size %d", len(metadataBlock))
		}

		if bytes.Equal(hash, metadataBlock) {
			continue
		}
		if _, err := metadataFile.Seek(-int64(read), io.SeekCurrent); err != nil {
			return fmt.Errorf("could not seek metadata file: %w", err)
		}

		n, err := metadataFile.Write(hash)
		if err != nil {
			return fmt.Errorf("could not write to metadata file: %w", err)
		}

		log.Printf("written %d bytes to metadata file %s", n, metadataFilePath)

		if !more {
			break
		}
	}

	pos, err := metadataFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	return metadataFile.Truncate(pos)
}

package hash

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
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

	// Leave space for the header in case the metadata file has changed
	_, err = metadataFile.Seek(8, io.SeekStart)
	if err != nil {
		return err
	}

	defer metadataFile.Close()

	// TODO: Metadata file should have a header with information about the last time it was written to it if changed
	// Note: The header part should also contain the previous timestamp of the update. Previous timestamp will be used to check changes between synchronization with the server.
	// Note: If the server has different last update time than the client's current previous update timestamp then there was clearly another update on the same file by some other process.
	// This should be handled with caution not to delete any changes but rather provide feedback that there are conflicting changes in the same file.

	/*
		File uploading should happen via partial files uploading. This is in case another device/process is trying to upload changes to the same file.
		There should be a process checking for finalized partial uploades. This process will handle the editing of the original file.
		If the process detects that there are 2 partial files

		version 1
		* Will not support handling of conflicts between updates from multiple devices
		* client should buffer changes for file during synchronization with server otherwise changes to metadata files might appear while uploading
		* client will save a local file with the last attemp at synchronization. If sync was successful will write to the actual file with the sync time started. If not successful nothing should change
		* client will fetch all metadata files edited after the last successful synchronization attempt
	*/

	changed := false

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

		_, err = metadataFile.Write(hash)
		if err != nil {
			return fmt.Errorf("could not write to metadata file: %w", err)
		}

		changed = true

		// log.Printf("written %d bytes to metadata file %s", n, metadataFilePath)

		if !more {
			break
		}
	}

	pos, err := metadataFile.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	if err := metadataFile.Truncate(pos); err != nil {
		return err
	}

	if changed {
		if _, err := metadataFile.Seek(0, io.SeekStart); err != nil {
			return nil
		}

		now := time.Now().UTC().Unix()
		byteArray := make([]byte, 8)

		binary.LittleEndian.PutUint64(byteArray, uint64(now))

		_, err = metadataFile.Write(byteArray)
		return err
	}

	return nil
}
